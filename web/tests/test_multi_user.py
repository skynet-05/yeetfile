import pytest
from playwright.sync_api import Browser, BrowserContext, Page, sync_playwright, expect

from constants import base_page, demo_file, user_password
from utils import fetch_folder_json, delete_account

account_id_a: str = ""
account_id_b: str = ""


@pytest.fixture(scope="session")
def browser_context() -> BrowserContext:
    with sync_playwright() as p:
        browser: Browser = p.chromium.launch(slow_mo=250)
        context_a: BrowserContext = browser.new_context()
        context_b: BrowserContext = browser.new_context()
        yield context_a, context_b
        browser.close()


def test_signup(browser_context: tuple[BrowserContext, BrowserContext]):
    """Creates two new id-only accounts"""
    global account_id_a, account_id_b
    context_a, context_b = browser_context

    def signup(page: Page) -> str:
        page.goto(f"{base_page}/signup")
        page.get_by_test_id("id-signup").click()

        signup_btn = page.get_by_test_id("create-id-only-account")
        expect(signup_btn).to_be_visible()

        page.get_by_test_id("account-password").fill(user_password)
        page.get_by_test_id("account-confirm-password").fill(user_password)

        signup_btn.click()
        expect(page.get_by_test_id("account-id-verify")).to_be_visible()

        page.get_by_test_id("account-code").fill("123456")
        page.get_by_test_id("verify-account").click()

        account_id = page.get_by_test_id("final-account-id").text_content()

        expect(page.get_by_test_id("goto-account")).to_be_visible()
        return account_id

    account_id_a = signup(context_a.new_page())
    account_id_b = signup(context_b.new_page())


def test_share_file(browser_context: tuple[BrowserContext, BrowserContext]):
    """User A uploads a file to their vault and shares it with User B"""
    global account_id_a, account_id_b
    context_a, context_b = browser_context

    file_content: str = "testing file sharing"
    with open(demo_file, "w") as f:
        f.write(file_content)

    user_a_page = context_a.new_page()
    user_a_page.goto(f"{base_page}/vault")

    user_a_page.get_by_test_id("file-input").set_input_files(demo_file)
    result_div = user_a_page.get_by_role("link", name=demo_file)
    result_div.wait_for()

    folder_json = fetch_folder_json(user_a_page, "")
    assert len(folder_json["items"]) == 1
    file_id = folder_json["items"][0]["id"]

    user_a_page.get_by_test_id(f"action-{file_id}").click()
    expect(user_a_page.get_by_test_id("actions-dialog")).to_be_visible()

    user_a_page.get_by_test_id("action-share").click()
    expect(user_a_page.get_by_test_id("share-dialog")).to_be_visible()

    user_a_page.get_by_test_id("share-target").fill(account_id_b)
    user_a_page.get_by_test_id("submit-share").click()

    user_b_page = context_b.new_page()
    user_b_page.goto(f"{base_page}/vault")

    file_link = user_b_page.get_by_role("link", name=demo_file)
    file_link.wait_for()


def test_share_folder(browser_context: tuple[BrowserContext, BrowserContext]):
    """User A creates a folder and shares it with User B"""
    global account_id_a, account_id_b
    context_a, context_b = browser_context

    folder_name = "My Folder"

    user_a_page = context_a.new_page()
    user_a_page.goto(f"{base_page}/vault")

    user_a_page.get_by_test_id("new-vault-folder").click()
    expect(user_a_page.get_by_test_id("folder-dialog")).to_be_visible()

    user_a_page.get_by_test_id("folder-name").fill(folder_name)
    user_a_page.get_by_test_id("submit-folder").click()
    result_div = user_a_page.get_by_role("link", name=folder_name)
    result_div.wait_for()

    folder_json = fetch_folder_json(user_a_page, "")
    assert len(folder_json["folders"]) == 1
    folder_id = folder_json["folders"][0]["id"]

    user_a_page.get_by_test_id(f"action-{folder_id}").click()
    expect(user_a_page.get_by_test_id("actions-dialog")).to_be_visible()

    user_a_page.get_by_test_id("action-share").click()
    expect(user_a_page.get_by_test_id("share-dialog")).to_be_visible()

    user_a_page.get_by_test_id("share-target").fill(account_id_b)
    user_a_page.get_by_test_id("submit-share").click()

    user_b_page = context_b.new_page()
    user_b_page.goto(f"{base_page}/vault")

    folder_link = user_b_page.get_by_role("link", name=folder_name)
    folder_link.wait_for()


def test_delete_accounts(browser_context: tuple[BrowserContext, BrowserContext]):
    global account_id_a, account_id_b
    context_a, context_b = browser_context

    user_a_page = context_a.new_page()
    user_b_page = context_b.new_page()

    delete_account(user_a_page, account_id_a)
    delete_account(user_b_page, account_id_b)

    expect(user_a_page).to_have_title("YeetFile - Send")
    expect(user_b_page).to_have_title("YeetFile - Send")
