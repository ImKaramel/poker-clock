import os
from dotenv import load_dotenv
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import ApplicationBuilder, CommandHandler, ContextTypes

load_dotenv()

BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
MINI_APP_URL = "https://midnight-club-app.ru"
API_BASE_URL = os.getenv("REACT_APP_API_URL", "https://api.midnight-club-app.ru/api")

if not BOT_TOKEN:
    raise ValueError("❌ TELEGRAM_BOT_TOKEN not set in .env")


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Команда /start — просто открываем Mini App"""

    user = update.effective_user

    keyboard = [[
        InlineKeyboardButton(
            "🎮 Открыть Midnight APP",
            web_app={"url": MINI_APP_URL}
        )
    ]]

    reply_markup = InlineKeyboardMarkup(keyboard)

    await update.message.reply_text(
                f"👋 Добро пожаловать в Midnight Club, {user.first_name or 'игрок'}!\n\n"
                "Нажмите кнопку ниже, чтобы открыть приложение!",
                reply_markup=reply_markup
            )



async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Команда /help"""
    help_text = (
        "🤖 Команды бота Poker CRM:\n\n"
        "/start - Начать работу с ботом\n"
        "/help - Показать эту справку\n\n"
        "📱 После запуска бота вы получите доступ к:\n"
        "• Расписанию турниров\n"
        "• Рейтингу игроков\n"
        "• Личному профилю\n"
        " Результатам игр\n\n"
        "По вопросам обращайтесь к администратору."
    )


def main():
    """Запуск Telegram-бота"""
    print("🤖 Запуск Telegram-бота Poker CRM...")


    app = ApplicationBuilder().token(BOT_TOKEN).build()

        # Добавляем обработчики команд
    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("help", help_command))

    print("✅ Бот успешно запущен. Ожидание сообщений...")
    app.run_polling()

if __name__ == "__main__":
    main()