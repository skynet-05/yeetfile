import pytest
import re
from playwright.sync_api import Browser, BrowserContext, Page, sync_playwright, expect

from constants import base_page, demo_file, user_password, file_password

from utils import fetch_folder_json, delete_account

account_id: str = ""


@pytest.fixture(scope="session")
def browser_context() -> BrowserContext:
    with sync_playwright() as p:
        browser: Browser = p.chromium.launch(slow_mo=200)
        context: BrowserContext = browser.new_context()
        yield context
        browser.close()


def test_has_title(browser_context: BrowserContext):
    page: Page = browser_context.new_page()
    page.goto(base_page)
    expect(page).to_have_title(re.compile(r"YeetFile - .*"))


def test_signup(browser_context: BrowserContext):
    """Creates a new id-only account"""
    global account_id
    page: Page = browser_context.new_page()
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


def test_logout(browser_context: BrowserContext):
    """Log out of YeetFile, ensuring access to other pages is blocked"""
    page: Page = browser_context.new_page()
    page.goto(f"{base_page}/account")
    page.on("dialog", lambda dialog: dialog.accept())
    page.get_by_test_id("logout-btn").click()
    expect(page).to_have_title("YeetFile - Send")

    page.goto(f"{base_page}/vault")
    expect(page).to_have_title("YeetFile - Log In")

    page.goto(f"{base_page}/account")
    expect(page).to_have_title("YeetFile - Log In")


def test_login(browser_context: BrowserContext):
    """Log back into YeetFile after logging out"""
    global account_id
    page: Page = browser_context.new_page()
    page.goto(f"{base_page}/login")
    page.get_by_test_id("identifier").fill(account_id)
    page.get_by_test_id("password").fill(user_password)
    page.get_by_test_id("login-btn").click()

    expect(page).to_have_title("YeetFile - My Account")


def test_text_send(browser_context: BrowserContext):
    """Test uploading text to YeetFile Send."""
    text_content: str = "testing text send"
    page: Page = browser_context.new_page()
    page.goto(f"{base_page}/send")

    page.get_by_test_id("upload-text-content").fill(text_content)
    page.get_by_test_id("downloads").fill("1")
    page.get_by_test_id("expiration").fill("5")
    page.get_by_test_id("submit").click()

    result_div = page.get_by_test_id("file-tag-div")
    result_div.wait_for()

    text_link = page.get_by_test_id("file-link").text_content()
    text_tag = page.get_by_test_id("file-tag").text_content()
    assert text_link.endswith(text_tag)

    page.goto(text_link)
    expect(page).to_have_title("YeetFile - Download")

    expect(page.get_by_test_id("password-prompt-div")).to_be_hidden()

    page.get_by_test_id("download-nopass").click()
    text_div = page.get_by_test_id("plaintext-div")
    text_div.wait_for()

    plaintext = page.get_by_test_id("plaintext-content").text_content()
    assert plaintext == text_content


def test_file_send(browser_context: BrowserContext):
    """Test uploading a file to YeetFile Send. This also sets a password
    for the file, unlike the text-only test."""
    file_content: str = "testing file send"
    with open(demo_file, "w") as f:
        f.write(file_content)

    page: Page = browser_context.new_page()
    page.goto(f"{base_page}/send")

    page.get_by_test_id("upload-file-btn").click()
    page.get_by_test_id("upload-file").set_input_files(demo_file)

    page.get_by_test_id("downloads").fill("2")
    page.get_by_test_id("expiration").fill("5")
    page.get_by_test_id("use-password").click()

    page.get_by_test_id("password").fill(file_password)
    page.get_by_test_id("confirm-password").fill(file_password)

    page.get_by_test_id("submit").click()

    result_div = page.get_by_test_id("file-tag-div")
    result_div.wait_for()

    text_link = page.get_by_test_id("file-link").text_content()
    text_tag = page.get_by_test_id("file-tag").text_content()
    assert text_link.endswith(text_tag)

    page.goto(text_link)
    expect(page).to_have_title("YeetFile - Download")
    submit_btn = page.get_by_test_id("submit")

    expect(page.get_by_test_id("password-prompt-div")).to_be_visible()
    expect(page.get_by_test_id("download-prompt-div")).to_be_hidden()

    page.get_by_test_id("password").fill("wrong password")
    submit_btn.click()

    # The wrong password was used, so the download prompt should still be hidden
    expect(page.get_by_test_id("download-prompt-div")).to_be_hidden()

    page.get_by_test_id("password").fill(file_password)
    submit_btn.click()

    # The correct password was used, so the download prompt should be visible
    expect(page.get_by_test_id("download-prompt-div")).to_be_hidden()
    expect(page.get_by_test_id("download-nopass")).to_be_visible()

    with page.expect_download() as download_info:
        page.get_by_test_id("download-nopass").click()

    download = download_info.value
    browser_name = browser_context.browser.browser_type.name
    new_filename = f"./test_{browser_name}.txt"
    download.save_as(new_filename)

    with open(new_filename) as f:
        assert f.readline() == file_content


def test_vault(browser_context: BrowserContext):
    """Test uploading a file to YeetFile Vault"""
    page: Page = browser_context.new_page()

    def _create_folder(folder_id: str) -> str:
        folder_name: str = "My Folder"
        page.goto(f"{base_page}/vault/{folder_id}")

        page.get_by_test_id("new-vault-folder").click()
        expect(page.get_by_test_id("folder-dialog")).to_be_visible()

        page.get_by_test_id("folder-name").fill(folder_name)
        page.get_by_test_id("submit-folder").click()

        expect(page.get_by_role("link", name=folder_name)).to_be_visible()

        folder_json = fetch_folder_json(page, folder_id)
        assert len(folder_json["folders"]) == 1
        return folder_json["folders"][0]["id"]

    def _upload_file(folder_id: str):
        file_content: str = "testing file vault"
        with open(demo_file, "w") as f:
            f.write(file_content)

        page.goto(f"{base_page}/vault/{folder_id}")

        page.get_by_test_id("file-input").set_input_files(demo_file)
        result_div = page.get_by_role("link", name=demo_file)
        result_div.wait_for()

        folder_json = fetch_folder_json(page, folder_id)
        assert len(folder_json["items"]) == 1
        file_id = folder_json["items"][0]["id"]

        page.get_by_test_id(f"action-{file_id}").click()
        expect(page.get_by_test_id("actions-dialog")).to_be_visible()

        with page.expect_download() as download_info:
            page.get_by_test_id("action-download").click()

        download = download_info.value
        browser_name = browser_context.browser.browser_type.name
        new_filename = f"./test_vault_{browser_name}.txt"
        download.save_as(new_filename)

        with open(new_filename) as new_f:
            assert new_f.readline() == file_content

        page.get_by_test_id(f"action-{file_id}").click()
        expect(page.get_by_test_id("actions-dialog")).to_be_visible()

        page.on("dialog", lambda dialog: dialog.accept())
        page.get_by_test_id("action-delete").click()
        expect(page.get_by_test_id("table-body")).to_be_empty()

        new_folder_json = fetch_folder_json(page, folder_id)
        assert len(new_folder_json["items"]) == 0

    _upload_file("")
    new_folder_id = _create_folder("")
    _upload_file(new_folder_id)


def test_vault_password(browser_context: BrowserContext):
    """Tests setting a unique session-specific vault password"""
    global account_id
    page: Page = browser_context.new_page()
    page.goto(f"{base_page}/account")
    page.on("dialog", lambda dialog: dialog.accept())
    page.get_by_test_id("logout-btn").click()
    expect(page).to_have_title("YeetFile - Send")

    page.goto(f"{base_page}/login")
    page.get_by_test_id("identifier").fill(account_id)
    page.get_by_test_id("password").fill(user_password)
    page.get_by_test_id("advanced-login-options").click()
    page.get_by_test_id("vault-pass-cb").click()
    page.get_by_test_id("login-btn").click()

    vault_password = "my_vault_password"
    expect(page.get_by_test_id("vault-pass-dialog")).to_be_visible()
    page.get_by_test_id("vault-pass").fill(vault_password)
    page.get_by_test_id("submit-pass").click()
    expect(page).to_have_title("YeetFile - My Account")

    page.goto(f"{base_page}/vault")
    expect(page.get_by_test_id("table-body")).to_be_empty()
    expect(page.get_by_test_id("vault-pass-dialog")).to_be_visible()
    page.get_by_test_id("vault-pass").fill("WRONG")
    page.get_by_test_id("submit-pass").click()
    expect(page.get_by_test_id("vault-pass-dialog")).to_be_visible()
    page.get_by_test_id("vault-pass").fill(vault_password)
    page.get_by_test_id("submit-pass").click()
    expect(page.get_by_test_id("vault-pass-dialog")).to_be_hidden()


def test_delete_account(browser_context: BrowserContext):
    """Permanently deletes the test user account"""
    global account_id
    page: Page = browser_context.new_page()
    delete_account(page, account_id)
    expect(page).to_have_title(re.compile(r"YeetFile - Send"))
