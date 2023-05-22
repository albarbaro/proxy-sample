describe('template spec', () => {
  it('passes', () => {
    cy.visit(Cypress.env('SPI_OAUTH_URL'))
    cy.window().then((win) => { 
      win.stop();
    });
    cy.origin('https://github.com/login', () => {
      cy.get('#login_field').type(Cypress.env('GH_USER'));
      cy.get('#password').type(Cypress.env('GH_PASSWORD'));
      cy.get('.btn').click();
    })
  })
})