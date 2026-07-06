/**
 * Structural Smoke Tests (CI-safe)
 *
 * These tests verify page structure and navigation without requiring S3 data.
 * They run against a local Next.js build with STORAGE_PROVIDER=local.
 */

describe('Home Page Structure', () => {
  it('renders the platform title', () => {
    cy.visit('/');
    cy.contains('PFRS Research Lab').should('be.visible');
  });

  it('renders 4 domain cards', () => {
    cy.visit('/');
    cy.contains('Nurse Rostering (NRP)').should('exist');
    cy.contains('Vehicle Routing (CVRP)').should('exist');
    cy.contains('Job Shop Scheduling (JSS)').should('exist');
    cy.contains('Vehicle Routing + Time Windows (VRPTW)').should('exist');
  });

  it('renders algorithm reference section', () => {
    cy.visit('/');
    cy.contains('Simulated Annealing').should('exist');
    cy.contains('Late Acceptance').should('exist');
    cy.contains('Tabu Search').should('exist');
    cy.contains('Portfolio Mode').should('exist');
    cy.contains('Adaptive Hyper-Heuristic').should('exist');
  });

  it('renders How It Works section', () => {
    cy.visit('/');
    cy.contains('How It Works').should('exist');
    cy.contains('1. Define Problem').should('exist');
    cy.contains('2. Run Algorithms').should('exist');
    cy.contains('3. Analyse Results').should('exist');
  });
});

describe('Sidebar Navigation', () => {
  it('shows all platform links', () => {
    cy.visit('/');
    cy.get('nav').within(() => {
      cy.contains('Home').should('exist');
      cy.contains('Benchmarks').should('exist');
      cy.contains('Statistics').should('exist');
      cy.contains('Compare').should('exist');
      cy.contains('Trends').should('exist');
      cy.contains('Admin').should('exist');
    });
  });

  it('navigates between pages', () => {
    cy.visit('/');
    cy.get('nav').contains('Benchmarks').click();
    cy.url().should('include', '/benchmarks');
  });
});

describe('Benchmarks Page Structure', () => {
  it('renders without errors', () => {
    cy.visit('/benchmarks');
    cy.get('body').should('not.contain', 'Error');
    cy.get('body').should('not.contain', 'error');
  });
});

describe('Statistics Page Structure', () => {
  it('renders without errors', () => {
    cy.visit('/statistics');
    cy.get('body').should('not.contain', 'Application error');
  });
});

describe('Admin Page Structure', () => {
  it('renders schema reference', () => {
    cy.visit('/admin');
    cy.contains('Platform Admin').should('be.visible');
    cy.contains('run.json Schema').should('exist');
    cy.contains('bestObjective').should('exist');
    cy.contains('problemType').should('exist');
    cy.contains('Objective Reading Priority').should('exist');
  });

  it('renders S3 storage layout', () => {
    cy.visit('/admin');
    cy.contains('S3 Storage Layout').should('exist');
    cy.contains('manifest.json').should('exist');
  });
});
