"""HiGHS LP/MIP solver script.

Called by the Go ILP benchmark tool. Reads an LP file, solves it,
writes JSON result to stdout. HiGHS native progress goes to stderr.
Also captures solve progress as a CSV for dashboard visualisation.

Usage:
    python solve_highs.py <model.lp> <time_limit_seconds> [progress_csv_path]
"""
import json
import sys
import time
import os
import re


def main():
    if len(sys.argv) < 3:
        emit_error("Usage: python solve_highs.py <model.lp> <time_limit> [progress_csv]")
        sys.exit(1)

    model_path = sys.argv[1]
    time_limit = float(sys.argv[2])
    progress_csv_path = sys.argv[3] if len(sys.argv) > 3 else None

    try:
        from highspy import Highs, HighsStatus
    except ImportError:
        emit_error("highspy not installed. Run: pip install highspy")
        sys.exit(1)

    h = Highs()
    h.setOptionValue("time_limit", time_limit)
    h.setOptionValue("parallel", "on")
    h.setOptionValue("threads", 16)

    # Direct HiGHS native logging to a temp file, then copy to stderr.
    log_path = model_path + ".log"
    h.setOptionValue("output_flag", True)
    h.setOptionValue("log_file", log_path)
    h.setOptionValue("log_to_console", False)

    start = time.time()

    status = h.readModel(model_path)
    if status not in (HighsStatus.kOk, HighsStatus.kWarning):
        emit_error(f"Failed to read model: {model_path}",
                   time.time() - start)
        sys.exit(1)

    num_cols = h.getNumCol()
    num_rows = h.getNumRow()
    print(f"Model: {num_cols} vars, {num_rows} constraints. Solving...",
          file=sys.stderr, flush=True)

    # Tail the log file in a background thread for real-time progress
    # and capture progress data points for the dashboard.
    import threading
    solving = [True]
    progress_points = []

    # Pattern to match HiGHS MIP progress lines, e.g.:
    # " B     1234    567    890  12.3%   1234.5          2345.6       10.1%    123   45   67    12345    12.3s"
    # Key columns: BestBound, BestSol, Gap%, LpIters, Time
    progress_pattern = re.compile(
        r'^\s*\S+\s+'           # Src
        r'(\d+)\s+'             # Proc (nodes processed)
        r'(\d+)\s+'             # InQueue
        r'(\d+)\s+'             # Leaves
        r'[\d.]+%\s+'           # Expl%
        r'([-\d.e+inf]+)\s+'   # BestBound
        r'([-\d.e+inf]+)\s+'   # BestSol
        r'([\d.]+)%\s+'        # Gap%
        r'(\d+)\s+'            # Cuts
        r'(\d+)\s+'            # InLp
        r'(\d+)\s+'            # Confl
        r'(\d+)\s+'            # LpIters
        r'([\d.]+)s'           # Time
    )

    def tail_log():
        """Tail the HiGHS log file, forward to stderr, and capture progress."""
        try:
            for _ in range(50):
                if os.path.exists(log_path):
                    break
                time.sleep(0.1)
            if not os.path.exists(log_path):
                return
            with open(log_path, "r") as f:
                while solving[0]:
                    line = f.readline()
                    if line:
                        sys.stderr.write(line)
                        sys.stderr.flush()
                        # Try to parse as a progress line.
                        m = progress_pattern.match(line)
                        if m:
                            try:
                                bound = parse_number(m.group(4))
                                incumbent = parse_number(m.group(5))
                                gap_pct = float(m.group(6))
                                nodes = int(m.group(1))
                                lp_iters = int(m.group(10))
                                elapsed = float(m.group(11))
                                progress_points.append({
                                    "elapsed": elapsed,
                                    "incumbent": incumbent,
                                    "bound": bound,
                                    "gap": gap_pct,
                                    "nodes": nodes,
                                    "lpIters": lp_iters,
                                })
                            except (ValueError, IndexError):
                                pass
                    else:
                        time.sleep(1)
        except Exception:
            pass

    t = threading.Thread(target=tail_log, daemon=True)
    t.start()

    h.run()
    solving[0] = False
    # Give tail thread a moment to finish reading.
    time.sleep(0.5)
    elapsed = time.time() - start

    # Clean up log file.
    try:
        os.remove(log_path)
    except Exception:
        pass

    model_status = h.getModelStatus()
    status_str = map_status(model_status, h)
    objective = get_objective(h)
    lower_bound = get_lower_bound(h, objective)
    variables = extract_variables(h, status_str)

    print(f"\nDone: status={status_str}, obj={objective:.1f}, "
          f"bound={lower_bound:.1f}, time={elapsed:.1f}s",
          file=sys.stderr, flush=True)

    # Write progress CSV if path provided.
    if progress_csv_path and progress_points:
        write_progress_csv(progress_csv_path, progress_points)
        print(f"Progress CSV: {len(progress_points)} data points written to {progress_csv_path}",
              file=sys.stderr, flush=True)

    result = {
        "status": status_str,
        "objective": safe_float(objective),
        "lowerBound": safe_float(lower_bound),
        "runtimeSeconds": round(elapsed, 3),
        "variables": variables,
        "progressPoints": len(progress_points),
    }
    print(json.dumps(result))


def parse_number(s):
    """Parse a number string, handling inf/-inf."""
    s = s.strip()
    if s in ('inf', '+inf'):
        return float('inf')
    if s == '-inf':
        return float('-inf')
    return float(s)


def write_progress_csv(path, points):
    """Write progress points to CSV."""
    os.makedirs(os.path.dirname(path) if os.path.dirname(path) else '.', exist_ok=True)
    with open(path, 'w') as f:
        f.write("elapsed,incumbent,bound,gap,nodes,lpIters\n")
        for p in points:
            inc = p['incumbent']
            bnd = p['bound']
            # Skip rows with inf values.
            if inc == float('inf') or inc == float('-inf'):
                inc = ""
            if bnd == float('inf') or bnd == float('-inf'):
                bnd = ""
            f.write(f"{p['elapsed']:.1f},{inc},{bnd},{p['gap']:.2f},{p['nodes']},{p['lpIters']}\n")


def map_status(model_status, h):
    from highspy import HighsStatus
    ms = str(model_status).upper()
    if "OPTIMAL" in ms:
        return "OPTIMAL"
    elif "INFEASIBLE" in ms:
        return "INFEASIBLE"
    elif "TIME_LIMIT" in ms or "OBJECTIVE_BOUND" in ms:
        return "FEASIBLE"
    try:
        info = h.getInfoValue("primal_solution_status")
        if isinstance(info, tuple) and info[1] == 2:
            return "FEASIBLE"
    except Exception:
        pass
    return "ERROR"


def get_objective(h):
    from highspy import HighsStatus
    try:
        v = h.getInfoValue("objective_function_value")
        if isinstance(v, tuple) and v[0] == HighsStatus.kOk:
            return v[1]
        elif isinstance(v, (int, float)):
            return float(v)
    except Exception:
        pass
    return 0.0


def get_lower_bound(h, fallback):
    from highspy import HighsStatus
    try:
        v = h.getInfoValue("mip_dual_bound")
        if isinstance(v, tuple) and v[0] == HighsStatus.kOk:
            return v[1]
        elif isinstance(v, (int, float)):
            return float(v)
    except Exception:
        pass
    return fallback


def extract_variables(h, status_str):
    from highspy import HighsStatus
    variables = {}
    if status_str not in ("OPTIMAL", "FEASIBLE"):
        return variables
    try:
        sol = h.getSolution()
        num_cols = h.getNumCol()
        col_names = []
        for i in range(num_cols):
            r = h.getColName(i)
            if isinstance(r, tuple) and len(r) >= 2:
                col_names.append(r[1])
            elif isinstance(r, str):
                col_names.append(r)
            else:
                col_names.append(f"c{i}")
        for i, val in enumerate(sol.col_value):
            if val > 0.5 and i < len(col_names):
                variables[col_names[i]] = round(val, 6)
    except Exception as e:
        print(f"Warning: could not extract solution: {e}",
              file=sys.stderr, flush=True)
    return variables


def safe_float(v):
    if v == float('inf') or v == float('-inf') or v != v:
        return 0.0
    return round(v, 6)


def emit_error(msg, elapsed=0):
    print(json.dumps({
        "status": "ERROR", "objective": 0, "lowerBound": 0,
        "runtimeSeconds": round(elapsed, 3), "variables": {},
        "error": msg,
    }))


if __name__ == "__main__":
    main()
