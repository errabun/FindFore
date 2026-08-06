// ***********************************************
// Cypress custom commands for API stubbing during e2e tests.
// Intercepts use the same origin as cypress.config.js baseUrl (local Vite dev server).
// ***********************************************

const apiBase = Cypress.config('baseUrl') || 'http://localhost:3000';

Cypress.Commands.add('setReadStubs', () => {
  cy.intercept(
    { method: 'GET', url: `${apiBase}/api/v1/players` },
    { fixture: '../fixtures/players.json' }
  );

  cy.intercept(
    { method: 'GET', url: `${apiBase}/api/v1/courses` },
    { fixture: '../fixtures/courses.json' }
  );

  cy.intercept(
    { method: 'GET', url: `${apiBase}/api/v1/players/1/events` },
    { fixture: '../fixtures/Events/initial.json' }
  ).as('getInitialEvents');
});

Cypress.Commands.add('setInviteActionStub', (action) => {
  cy.wait('@getInitialEvents');

  cy.intercept(
    { method: 'GET', url: `${apiBase}/api/v1/players/1/events` },
    { fixture: `../fixtures/Events/after_${action}.json` }
  );
});

Cypress.Commands.add('setUpdateStub', () => {
  cy.intercept('PATCH', `${apiBase}/api/v1/player-event`, {});
});

Cypress.Commands.add('setDeleteStub', () => {
  cy.intercept('DELETE', `${apiBase}/api/v1/event/1`, {});
});

Cypress.Commands.add('setFriendshipStub', () => {
  cy.intercept({
    method: 'POST',
    url: `${apiBase}/api/v1/friendship`,
  });
  cy.intercept({
    method: 'DELETE',
    url: `${apiBase}/api/v1/friendship`,
  }, {});
});
