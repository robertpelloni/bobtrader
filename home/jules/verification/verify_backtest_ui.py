from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    # Assuming the frontend is running on localhost:5173 (standard Vite port)
    # The backend should be running on localhost:3000
    try:
        page.goto("http://localhost:5173/strategy-sandbox")

        # Wait for page load
        page.wait_for_selector("h1:has-text('Strategy Sandbox')")

        # Select Strategy (should be populated from backend)
        # We might need to wait for the select to populate if it fetches from API
        page.wait_for_timeout(2000) # Give a moment for strategies to load

        # Click Run Backtest
        page.click("button:has-text('Run Backtest')")

        # Wait for results
        # We look for "Results" header which appears conditionally
        page.wait_for_selector("h3:has-text('Results')", timeout=10000)

        # Wait a bit for charts to render
        page.wait_for_timeout(2000)

        # Take Screenshot
        page.screenshot(path="/home/jules/verification/backtest_ui.png", full_page=True)
        print("Screenshot saved to /home/jules/verification/backtest_ui.png")

    except Exception as e:
        print(f"Verification failed: {e}")
        page.screenshot(path="/home/jules/verification/error.png")
    finally:
        browser.close()

with sync_playwright() as playwright:
    run(playwright)
