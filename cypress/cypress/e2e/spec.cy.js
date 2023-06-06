describe('template spec', () => {
  it('passes', () => {
    cy.log(Cypress.env('SPI_OAUTH_URL'))
    cy.log(Cypress.env('GH_USER'))
    cy.log(Cypress.env('GH_PASSWORD'))
    
    cy.visit(Cypress.env('SPI_OAUTH_URL'))

    cy.origin('https://github.com/login', () => {
      cy.get('#login_field').should('exist')
      cy.get('#password').should('exist')
      cy.get('input[type="submit"][name="commit"]').should('exist')

      cy.get('#login_field').type(Cypress.env('GH_USER'));
      cy.get('#password').type(Cypress.env('GH_PASSWORD'));
      cy.get('input[type="submit"][name="commit"]').click();
      
      cy.get('body').then(($el) => {
        if ($el.find('#js-oauth-authorize-btn').length > 0) {
          cy.log('Need to authorize app')
          cy.get('#js-oauth-authorize-btn').click();
        } else {
          cy.log('No need to authorize app')
        }
      });
    })

    /*cy.location('pathname')
      .should('include', '/callback_success')
      .then(cy.log)*/
    
    /*cy.task("generateToken", "some").then(token => {
      cy.get("#app_totp").type(token);
    });*/
  })
})

