import asyncio
import logging
import os
from typing import Any

import requests
from dotenv import load_dotenv
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, Update
from telegram.error import BadRequest, Forbidden, RetryAfter, TelegramError
from telegram.ext import (
    ApplicationBuilder,
    CallbackQueryHandler,
    CommandHandler,
    ContextTypes,
    MessageHandler,
    filters,
)

load_dotenv()

logging.basicConfig(
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    level=logging.INFO,
)
logger = logging.getLogger(__name__)

BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
API_BASE_URL = os.getenv("API_BASE_URL", "https://api.midnight-club-app.ru/api")
ADMIN_TELEGRAM_IDS = {
    value.strip()
    for value in os.getenv("ADMIN_TELEGRAM_IDS", "").split(",")
    if value.strip()
}

PROMOTIONS_URL = "https://t.me/midnight_poker_club/77"
ADMIN_URL = "https://t.me/midnight_club_admin"
LADIES_BONUS_TEXT = "Девушкам всегда — первый вход FREE! (Кроме FREEZEOUT)"
LADIES_BONUS_CALLBACK = "ladies_bonus"
AWAITING_BROADCAST_KEY = "awaiting_broadcast"
BOT_TOKEN_HEADER = "X-Bot-Token"

if not BOT_TOKEN:
    raise ValueError("❌ TELEGRAM_BOT_TOKEN not set in .env")


def is_admin(update: Update) -> bool:
    user = update.effective_user
    return user is not None and str(user.id) in ADMIN_TELEGRAM_IDS


def broadcast_markup() -> InlineKeyboardMarkup:
    keyboard = [
        [
            InlineKeyboardButton("Акции", url=PROMOTIONS_URL),
            InlineKeyboardButton("Ссылка на администратора", url=ADMIN_URL),
        ],
        [
            InlineKeyboardButton("Бонус дамам", callback_data=LADIES_BONUS_CALLBACK),
        ],
    ]
    return InlineKeyboardMarkup(keyboard)


def fetch_recipients() -> list[dict[str, Any]]:
    response = requests.get(
        f"{API_BASE_URL}/bot/recipients",
        headers={BOT_TOKEN_HEADER: BOT_TOKEN},
        timeout=20,
    )
    response.raise_for_status()
    payload = response.json()
    users = payload.get("users", [])
    if not isinstance(users, list):
        raise ValueError("invalid recipients payload")
    return users


def parse_broadcast_text(args: list[str]) -> str:
    return " ".join(args).strip()


async def ensure_admin(update: Update) -> bool:
    if not ADMIN_TELEGRAM_IDS:
        if update.effective_message:
            await update.effective_message.reply_text(
                "Для рассылки не настроены ADMIN_TELEGRAM_IDS. Добавим их в env и всё заработает."
            )
        return False

    if not is_admin(update):
        if update.effective_message:
            await update.effective_message.reply_text("Эта команда доступна только администратору.")
        return False
    return True


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user = update.effective_user
    lines = [
        f"👋 Добро пожаловать в Midnight Club, {user.first_name or 'игрок'}!",
        "",
        "Я могу присылать напоминания и важные анонсы по турнирам.",
        "Команды:",
        "/start — приветствие",
        "/help — список команд",
    ]
    if is_admin(update):
        lines.extend(
            [
                "/broadcast <текст> — отправить рассылку сразу",
                "/broadcast — перейти в режим ввода текста",
                "/cancel_broadcast — отменить режим рассылки",
            ]
        )

    await update.effective_message.reply_text("\n".join(lines))


async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    help_lines = [
        "🤖 Команды бота Midnight Club:",
        "",
        "/start — начать работу",
        "/help — показать эту справку",
    ]
    if is_admin(update):
        help_lines.extend(
            [
                "",
                "Админ-рассылка:",
                "/broadcast <текст> — отправить сообщение всем пользователям бота",
                "/broadcast — попросить текст следующим сообщением",
                "/cancel_broadcast — отменить рассылку",
            ]
        )
    await update.effective_message.reply_text("\n".join(help_lines))


async def broadcast_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await ensure_admin(update):
        return

    text = parse_broadcast_text(context.args)
    if text:
        context.user_data.pop(AWAITING_BROADCAST_KEY, None)
        await run_broadcast(update, context, text)
        return

    context.user_data[AWAITING_BROADCAST_KEY] = True
    await update.effective_message.reply_text(
        "Отправь следующим сообщением текст рассылки, и я разошлю его всем пользователям."
    )


async def cancel_broadcast(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await ensure_admin(update):
        return

    if context.user_data.pop(AWAITING_BROADCAST_KEY, None):
        await update.effective_message.reply_text("Ок, режим рассылки отменён.")
        return

    await update.effective_message.reply_text("Сейчас активной рассылки на ввод нет.")


async def handle_broadcast_text(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not context.user_data.get(AWAITING_BROADCAST_KEY):
        return

    if not await ensure_admin(update):
        context.user_data.pop(AWAITING_BROADCAST_KEY, None)
        return

    text = (update.effective_message.text or "").strip()
    if not text:
        await update.effective_message.reply_text("Текст пустой. Пришли обычное текстовое сообщение.")
        return

    context.user_data.pop(AWAITING_BROADCAST_KEY, None)
    await run_broadcast(update, context, text)


async def ladies_bonus_callback(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    query = update.callback_query
    if query is None:
        return
    await query.answer(LADIES_BONUS_TEXT, show_alert=True)


async def send_broadcast_message(
    context: ContextTypes.DEFAULT_TYPE,
    chat_id: int,
    text: str,
) -> None:
    await context.bot.send_message(
        chat_id=chat_id,
        text=text,
        reply_markup=broadcast_markup(),
    )


async def run_broadcast(
    update: Update,
    context: ContextTypes.DEFAULT_TYPE,
    text: str,
) -> None:
    status_message = await update.effective_message.reply_text("Собираю получателей и начинаю рассылку.")

    try:
        recipients = fetch_recipients()
    except Exception as exc:  # noqa: BLE001
        logger.exception("failed to fetch recipients")
        await status_message.edit_text(f"Не удалось получить список получателей: {exc}")
        return

    sent = 0
    failed = 0
    skipped = 0

    for recipient in recipients:
        raw_chat_id = str(recipient.get("user_id", "")).strip()
        if not raw_chat_id or not raw_chat_id.lstrip("-").isdigit():
            skipped += 1
            continue

        chat_id = int(raw_chat_id)
        try:
            await send_broadcast_message(context, chat_id, text)
            sent += 1
        except RetryAfter as exc:
            await asyncio.sleep(int(exc.retry_after) + 1)
            try:
                await send_broadcast_message(context, chat_id, text)
                sent += 1
            except (Forbidden, BadRequest, TelegramError):
                failed += 1
        except (Forbidden, BadRequest, TelegramError):
            failed += 1

        await asyncio.sleep(0.05)

    await status_message.edit_text(
        "Рассылка завершена.\n"
        f"Успешно: {sent}\n"
        f"Ошибки: {failed}\n"
        f"Пропущено: {skipped}"
    )


def main() -> None:
    logger.info("🤖 Запуск Telegram-бота Midnight Club...")
    if not ADMIN_TELEGRAM_IDS:
        logger.warning("ADMIN_TELEGRAM_IDS is empty: broadcast commands will be disabled")

    app = ApplicationBuilder().token(BOT_TOKEN).build()

    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("help", help_command))
    app.add_handler(CommandHandler("broadcast", broadcast_command))
    app.add_handler(CommandHandler("cancel_broadcast", cancel_broadcast))
    app.add_handler(CallbackQueryHandler(ladies_bonus_callback, pattern=f"^{LADIES_BONUS_CALLBACK}$"))
    app.add_handler(
        MessageHandler(
            filters.TEXT & ~filters.COMMAND,
            handle_broadcast_text,
        )
    )

    logger.info("✅ Бот запущен. Ожидание сообщений...")
    app.run_polling()


if __name__ == "__main__":
    main()
