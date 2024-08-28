from constants import base_page
from playwright.sync_api import Locator, Page


def fetch_folder_json(page: Page, folder_id: str) -> dict:
    json_data = page.evaluate(f"""
        async () => {{
            const response = await fetch("/api/vault/folder/{folder_id}");
            const data = await response.json();
            return data;
        }}
    """)

    return json_data


def delete_account(page: Page, account_id: str):
    page.goto(f"{base_page}/account")

    page.get_by_test_id("advanced-summary").click()
    page.get_by_test_id("delete-btn").click()

    page.on("dialog", lambda dialog: dialog.accept(account_id))
    page.get_by_test_id("delete-btn").click()
