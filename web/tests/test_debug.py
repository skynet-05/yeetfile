import pytest
from playwright.sync_api import Browser, BrowserContext, Page, sync_playwright

from constants import base_page


@pytest.fixture(scope="session")
def browser_context() -> BrowserContext:
    with sync_playwright() as p:
        browser: Browser = p.chromium.launch(headless=False)
        context: BrowserContext = browser.new_context()
        yield context
        browser.close()


def test_debug(browser_context: BrowserContext):
    page: Page = browser_context.new_page()
    page.goto(base_page)
    page.wait_for_timeout(5000000)
