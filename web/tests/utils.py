from constants import base_page
from playwright.sync_api import Locator, Page, ConsoleMessage


def fetch_folder_json(page: Page, folder_id: str, pass_vault: bool = False) -> dict:
    api = "/api/vault/folder"
    if pass_vault:
        api = "/api/pass/folder"

    json_data = page.evaluate(f"""
        async () => {{
            const response = await fetch("{api}/{folder_id}");
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


def print_console_msg(msg: ConsoleMessage):
    print(f'\nconsole.{msg.type}: {msg.text}')
    if len(msg.args) > 0:
        for idx, arg in enumerate(msg.args):
            print(f'Argument {idx}: {arg.json_value()}')
