"""HiGHS LP/MIP solver script.

Called by the Go ILP benchmark tool. Reads an LP file, solves it,
writes JSON result to stdout. HiGHS native progress goes to stderr.

Usage:
    python solve_highs.py <model.lp> <time_limit_seconds>
"""
import json
import sys
import time
import os


def main():
    if len(sys.argv) < 3:
        emit_error("Usage: python solve_highs.py <model.lp> <time_limit>")
        sys.exit(1)

    model_path = sys.argv[1]
    time_limit = float(sys.argv[2])

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
    # This gives real progress (gap, nodes, incumbent) without corrupting stdout JSON.
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

    # Tail the log file in a background thread for real-time progress.
    import threading
    solving = [True]

    def tail_log():
        """Tail the HiGHS log file and forward to stderr."""
        try:
            # Wait for file to be created.
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
                    else:
                        time.sleep(1)
        except Exception:
            pass

    t = threading.Thread(target=tail_log, daemon=True)
    t.start()

    h.run()
    solving[0] = False
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

    result = {
        "status": status_str,
        "objective": safe_float(objective),
        "lowerBound": safe_float(lower_bound),
        "runtimeSeconds": round(elapsed, 3),
        "variables": variables,
    }
    print(json.dumps(result))


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
