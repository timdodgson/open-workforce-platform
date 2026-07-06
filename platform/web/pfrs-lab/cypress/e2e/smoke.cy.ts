/**
 * Dashboard Smoke Tests
 *
 * Verifies that all key pages load and display expected content.
 * Runs against the deployed dashboard (CYPRESS_BASE_URL) or localhost.
 */

describe('Home Page', () => {
  it('loads and shows platform title', () => {
    cy.visit('/');
    cy.contains('PFRS Research Lab').should('be.visible');
  });

  it('shows all 4 domain cards', () => {
    cy.visit('/');
    cy.contains('Nurse Rostering (NRP)').should('exist');
    cy.contains('Vehicle Routing (CVRP)').should('exist');
    cy.contains('Job Shop Scheduling (JSS)').should('exist');
    cy.contains('Vehicle Routing + Time Windows (VRPTW)').should('exist');
  });

  it('shows all 5 algorithms', () => {
    cy.visit('/');
    cy.contains('Simulated Annealing').should('exist');
    cy.contains('Late Acceptance').should('exist');
    cy.contains('Tabu Search').should('exist');
    cy.contains('Portfolio Mode').should('exist');
    cy.contains('Adaptive Hyper-Heuristic').should('exist');
  });
});

describe('Benchmarks Page', () => {
  it('loads and shows leaderboard', () => {
    cy.visit('/benchmarks');
    cy.contains('Algorithm Leaderboard').should('be.visible');
  });

  it('shows experimental setup', () => {
    cy.visit('/benchmarks');
    cy.contains('Experimental Setup').should('be.visible');
    cy.contains('ALGORITHMS').should('exist');
    cy.contains('ITERATIONS').should('exist');
  });

  it('shows CVRP section', () => {
    cy.visit('/benchmarks');
    cy.contains('Vehicle Routing (CVRPLIB)').should('exist');
  });

  it('shows JSS section', () => {
    cy.visit('/benchmarks');
    cy.contains('Job Shop Scheduling (Taillard)').should('exist');
  });

  it('shows platform summary', () => {
    cy.visit('/benchmarks');
    cy.contains('Platform Summary').should('exist');
    cy.contains('PROBLEM DOMAINS').should('exist');
  });
});

describe('Statistics Page', () => {
  it('loads and shows domain filter', () => {
    cy.visit('/statistics');
    cy.contains('Statistical Analysis').should('be.visible');
    cy.contains('CVRP').should('exist');
    cy.contains('JSS').should('exist');
    cy.contains('NRP').should('exist');
  });

  it('shows box plots', () => {
    cy.visit('/statistics');
    cy.contains('Distribution (Box Plots)').should('exist');
  });
});

describe('Admin Page', () => {
  it('loads and shows system info', () => {
    cy.visit('/admin');
    cy.contains('Platform Admin').should('be.visible');
    cy.contains('STORAGE').should('exist');
    cy.contains('TOTAL RUNS').should('exist');
  });

  it('shows run.json schema', () => {
    cy.visit('/admin');
    cy.contains('run.json Schema').should('exist');
    cy.contains('bestObjective').should('exist');
    cy.contains('problemType').should('exist');
  });

  it('shows objective reading priority', () => {
    cy.visit('/admin');
    cy.contains('Objective Reading Priority').should('exist');
  });
});

describe('Run Summary - CVRP', () => {
  it('loads CVRP run and shows distance', () => {
    cy.visit('/runs/cvrp-a45k6-sa/summary');
    cy.contains('Run: cvrp-a45k6-sa').should('be.visible');
    cy.contains('BEST DISTANCE').should('exist');
    cy.contains('FEASIBLE').should('exist');
  });
});

describe('Run Summary - JSS', () => {
  it('loads JSS run and shows makespan', () => {
    cy.visit('/runs/jss-ft06-sa/summary');
    cy.contains('Run: jss-ft06-sa').should('be.visible');
    cy.contains('BEST MAKESPAN').should('exist');
  });
});

describe('Run Summary - VRPTW', () => {
  it('loads VRPTW run and shows vehicles', () => {
    cy.visit('/runs/vrptw-c101-sa/summary');
    cy.contains('Run: vrptw-c101-sa').should('be.visible');
    cy.contains('VEHICLES USED').should('exist');
    cy.contains('BEST DISTANCE').should('exist');
  });
});

describe('Gantt Chart - JSS', () => {
  it('loads and shows chart with operations', () => {
    cy.visit('/runs/jss-ft06-sa/gantt');
    cy.contains('Job Shop Schedule (Gantt Chart)').should('be.visible');
    cy.contains('Makespan:').should('exist');
    cy.contains('Operations:').should('exist');
    // Verify SVG elements exist (operations rendered).
    cy.get('svg rect').should('have.length.greaterThan', 10);
  });
});

describe('Navigation', () => {
  it('sidebar shows correct links', () => {
    cy.visit('/');
    cy.contains('Home').should('exist');
    cy.contains('Benchmarks').should('exist');
    cy.contains('Statistics').should('exist');
    cy.contains('Admin').should('exist');
  });

  it('navigates to benchmarks', () => {
    cy.visit('/');
    cy.contains('Benchmarks').click();
    cy.url().should('include', '/benchmarks');
    cy.contains('Algorithm Leaderboard').should('be.visible');
  });
});
