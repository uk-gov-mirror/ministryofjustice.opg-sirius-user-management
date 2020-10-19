describe("Add a user to a team", () => {
    beforeEach(() => {
        cy.setCookie("Other", "other");
        cy.setCookie("XSRF-TOKEN", "abcde");
        cy.visit("/teams/add-member/65");
    });

    it("shows search button", () => {
        cy.contains(".govuk-button", "Search")
    });
    
});
