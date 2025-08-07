import re
from playwright.sync_api import sync_playwright, Page, expect

def verify_profile_page(page: Page):
    """
    This script verifies that a user can log in, navigate to the profile page,
    and see the profile form.
    """
    # 1. Navigate to the login page.
    page.goto("http://localhost:3000/admin/login")

    # 2. Fill in the credentials.
    page.get_by_label("Email Address").fill("admin@gorectus.local")
    page.get_by_label("Password").fill("admin123")

    # 3. Click the login button.
    page.get_by_role("button", name="Sign In").click()

    # 4. Wait for navigation to the dashboard.
    expect(page).to_have_url(re.compile(r".*/dashboard"))

    # 5. Click the avatar to open the profile menu.
    page.get_by_label("account of current user").click()

    # 6. Click the "Profile" link.
    page.get_by_role("menuitem", name="Profile").click()

    # 7. Wait for the profile page to load.
    expect(page).to_have_url(re.compile(r".*/profile"))
    expect(page.get_by_role("heading", name="My Profile")).to_be_visible()

    # 8. Take a screenshot.
    page.screenshot(path="jules-scratch/verification/profile_page.png")

def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        verify_profile_page(page)
        browser.close()

if __name__ == "__main__":
    main()
