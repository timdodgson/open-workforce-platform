/**
 * Structural Smoke Tests (CI-safe)
 *
 * These tests verify page structure and navigation without requiring S3 data.
 * They run against a local Next.js build with STORAGE_PROVIDER=local.
 */

describe('Public Home Page', () => {
  it('renders the platform title', () => {
    cy.visit('/');
    cy.contains('PFRS Lab').should('be.visible');
  });

  it('renders marketing hero and proof metrics', () => {
    cy.visit('/');
    cy.contains('Hard problems').should('exist');
    cy.contains('benchmark domains').should('exist');
    cy.contains('Quick start').should('exist');
    cy.contains('Explore the live lab').should('exist');
    cy.contains('nurse rostering').should('exist');
  });

  it('renders audience sections', () => {
    cy.visit('/');
    cy.contains('Researchers').should('exist');
    cy.contains('University students').should('exist');
    cy.contains('Cite & reproduce').should('exist');
    cy.contains('Learning path').should('exist');
    cy.contains('Quick start').should('exist');
    cy.contains('About').should('exist');
  });

  it('renders field guide cards', () => {
    cy.visit('/');
    cy.contains('Field guide').should('exist');
    cy.contains('Algorithms').should('exist');
    cy.contains('Domains').should('exist');
    cy.contains('References').should('exist');
  });

  it('renders BYOD extension teaser', () => {
    cy.visit('/');
    cy.contains('Extend with your own domain').should('exist');
  });
});

describe('Getting Started Page', () => {
  it('leads with Quick start and expected CVRP distance', () => {
    cy.visit('/getting-started');
    cy.contains('Quick start').should('exist');
    cy.contains('784').should('exist');
    cy.contains('Switch reference').should('exist');
  });
});

describe('Field Guide Pages', () => {
  it('renders algorithms page', () => {
    cy.visit('/algorithms');
    cy.contains('Genetic').should('exist');
    cy.contains('Portfolio').should('exist');
  });

  it('renders domains page', () => {
    cy.visit('/domains');
    cy.contains('Nurse Rostering').should('exist');
  });
});

describe('Public Site Navigation', () => {
  it('shows marketing header links', () => {
    cy.visit('/');
    cy.get('header nav').within(() => {
      cy.contains('Get started').should('be.visible');
      cy.contains('Algorithms').should('be.visible');
      cy.contains('Domains').should('be.visible');
      cy.contains('References').should('be.visible');
      cy.contains('Research').should('be.visible');
      cy.contains('Reproduce').should('be.visible');
      cy.contains('About').should('be.visible');
    });
    cy.contains('Open Lab').should('exist');
  });

  it('navigates to research depth page', () => {
    cy.visit('/');
    cy.get('header nav').contains('Research').click();
    cy.url().should('include', '/research');
  });

  it('navigates to lab from CTA', () => {
    cy.visit('/');
    cy.contains('Explore the live lab').click();
    cy.url().should('include', '/lab');
  });
});

describe('Reproduce Page Structure', () => {
  it('renders student path and citation blocks', () => {
    cy.visit('/reproduce');
    cy.contains('Cite & reproduce').should('exist');
    cy.contains('Student learning path').should('exist');
    cy.contains('Level 1').should('exist');
    cy.contains('EXP-008').should('exist');
    cy.contains('BibTeX').should('exist');
  });

  it('uses public site shell without lab sidebar', () => {
    cy.visit('/reproduce');
    cy.get('nav.w-56').should('not.exist');
    cy.get('header nav').contains('Reproduce').should('exist');
  });
});

describe('Research Page Structure', () => {
  it('renders 4 domain cards', () => {
    cy.visit('/research');
    cy.contains('NRP').should('exist');
    cy.contains('CVRP').should('exist');
    cy.contains('JSS').should('exist');
    cy.contains('VRPTW').should('exist');
    cy.contains('Nurse Rostering').should('exist');
    cy.contains('Vehicle Routing').should('exist');
    cy.contains('Job Shop').should('exist');
  });

  it('renders algorithm reference section', () => {
    cy.visit('/research');
    cy.contains('Simulated Annealing').should('exist');
    cy.contains('Late Acceptance').should('exist');
    cy.contains('Tabu Search').should('exist');
    cy.contains('Portfolio').should('exist');
    cy.contains('Search Intelligence').should('exist');
  });

  it('renders Search Intelligence modes', () => {
    cy.visit('/research');
    cy.contains('Search Intelligence').should('exist');
    cy.contains('off').should('exist');
    cy.contains('shadow').should('exist');
    cy.contains('assist').should('exist');
    cy.contains('adaptive').should('exist');
  });
});

describe('Lab Sidebar Navigation', () => {
  it('shows lab platform links', () => {
    cy.visit('/lab');
    cy.get('header').within(() => {
      cy.contains('Get started').should('exist');
      cy.contains('Reproduce').should('exist');
      cy.contains('About').should('exist');
    });
    cy.get('nav.w-56').within(() => {
      cy.contains('Lab Home').should('exist');
      cy.contains('Getting Started').should('exist');
      cy.contains('About').should('exist');
      cy.contains('Benchmarks').should('exist');
      cy.contains('Statistics').should('exist');
      cy.contains('Compare').should('exist');
      cy.contains('Trends').should('exist');
      cy.contains('Admin').should('exist');
    });
  });

  it('navigates between lab pages', () => {
    cy.visit('/lab');
    cy.get('nav.w-56').contains('Benchmarks').click();
    cy.url().should('include', '/benchmarks');
  });
});

describe('Benchmarks Page Structure', () => {
  it('renders without application error', () => {
    cy.visit('/benchmarks');
    cy.get('body').should('not.contain', 'Application error');
    cy.get('body').should('not.contain', 'Internal Server Error');
    // Should show either the ladder or the empty state.
    cy.get('body').then(($body) => {
      const text = $body.text();
      const hasLadder = text.includes('Algorithm Leaderboard');
      const hasEmpty = text.includes('No benchmark data');
      expect(hasLadder || hasEmpty).to.be.true;
    });
  });
});

describe('Statistics Page Structure', () => {
  it('renders without application error', () => {
    cy.visit('/statistics');
    cy.get('body').should('not.contain', 'Application error');
    cy.get('body').should('not.contain', 'Internal Server Error');
  });
});

describe('Compare Page Structure', () => {
  it('renders without application error', () => {
    cy.visit('/compare');
    cy.get('body').should('not.contain', 'Application error');
    cy.get('body').should('not.contain', 'Internal Server Error');
    // Should show either the comparison UI or the empty state.
    cy.get('body').then(($body) => {
      const text = $body.text();
      const hasCompare = text.includes('Head-to-Head') || text.includes('Compare');
      const hasEmpty = text.includes('Need at least 2 runs');
      expect(hasCompare || hasEmpty).to.be.true;
    });
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
