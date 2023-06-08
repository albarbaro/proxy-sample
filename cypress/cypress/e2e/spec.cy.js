describe('template spec', () => {
  it('passes', () => {   
    if(Cypress.env('SPI_OAUTH_URL')==""){
      cy.task('log', 'SPI_OAUTH_URL is empty')
      throw new Error("SPI_OAUTH_URL is empty")
    }
    if(Cypress.env('GH_USER')==""){
       cy.task('log', 'GH_USER is empty')
      throw new Error("GH_USER is empty")
    }
    if(Cypress.env('GH_PASSWORD')==""){
       cy.task('log', 'GH_PASSWORD is empty')
      throw new Error("GH_PASSWORD is empty")
    }
    if(Cypress.env('GH_2FA_CODE')==""){
       cy.task('log', 'GH_2FA_CODE is empty')
      throw new Error("GH_2FA_CODE is empty")
    }
    
    cy.task('log', 'Visiting '+Cypress.env('SPI_OAUTH_URL'))
    cy.visit(Cypress.env('SPI_OAUTH_URL'))
    
    cy.url().then((url) => {
      cy.task('log', 'Current URL is: ' + url)
    })

    cy.origin('https://github.com/login', () => {
      cy.url().then((url) => {
        cy.task('log', 'Current Github URL is: ' + url)
      })
      
      cy.get('#login_field').should('exist')
      cy.get('#password').should('exist')
      cy.get('input[type="submit"][name="commit"]').should('exist')

      cy.get('#login_field').type(Cypress.env('GH_USER'));
      cy.get('#password').type(Cypress.env('GH_PASSWORD'));
      cy.get('input[type="submit"][name="commit"]').click();
      
      cy.task("generateToken", Cypress.env('GH_2FA_CODE')).then(token => {
        cy.get("#app_totp").type(token);
        cy.task('log', 'Generated token')
        if(token==""){
          cy.task('log', 'token is empty')
          throw new Error("token is empty")
        }
      });
      
      cy.get('body').then(($el) => {
        if ($el.find('#js-oauth-authorize-btn').length > 0) {
          cy.task('log', 'Need to authorize app')
          cy.get('#js-oauth-authorize-btn').click();
        } else {
          cy.task('log', 'No need to authorize app')
        }
      });
    })
    
    cy.location('pathname')
      .then((url) => {
        cy.task('log', 'Current URL is: ' + url)
      })
      .should('include', '/callback_success')
    
    cy.url().then((url) => {
      cy.task('log', 'Current URL is: ' + url)
    })
  })
})

