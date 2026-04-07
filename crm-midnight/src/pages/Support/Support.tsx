import React, { useState } from "react";
import {
  SupportContainer,
  SupportTitle,
  FaqList,
  FaqItem,
  QuestionButton,
  QuestionText,
  ArrowIcon,
  AnswerContainer,
  AnswerText,
  TelegramButton,
} from "./styles";

export default function Support() {
  const [openIndex, setOpenIndex] = useState(null);

  const faqData = [
    {
      question: "Это законно?",
      answer:
        "Да, наша платформа полностью соответствует законодательству. Мы работаем в рамках правового поля и регулярно проходим все необходимые проверки. Все турниры проводятся в соответствии с установленными правилами и нормами.",
    },
    {
      question: "Как записаться на игру?",
      answer:
        "Записаться на игры можно прямо в нашем приложении на вкладке «Турниры». Выбирайте любые турнирные форматы, которые зажгут вашу душу, - от эпических баталий до хитрых стратегий!",
    },
    {
      question: "Сколько стоит участие?",
      answer:
        "Информация об организационном сборе указывается в карточке турнира",
    },
    {
      question: "Что такое стартовый стек?",
      answer:
        "Стартовый стек — это начальное количество фишек, которое получает каждый участник в начале турнира. Размер стартового стека зависит от формата турнира и может варьироваться. Чем больше стартовый стек, тем больше пространства для манёвра и стратегии в начале игры.",
    },
    {
      question: "Зачем играть без призов?",
      answer:
        "Игры без призов — это отличная возможность отточить свои навыки, протестировать новые стратегии, завести новые знакомства и просто получить удовольствие от игры без давления",
    },
    {
      question: "Что такое рейтинг?",
      answer:
        "Рейтинг — это система оценки вашего мастерства и достижений в турнирах. Чем успешнее вы выступаете, тем выше становится ваш рейтинг. Следите за своим рейтингом и стремитесь к вершине!",
    },
    {
      question: "Что такое финал месяца?",
      answer:
        "Финал месяца — это главное ежемесячное событие нашей платформы, в котором участвуют лучшие игроки по итогам месяца. Чтобы попасть в финал, нужно набрать достаточное количество очков рейтинга в течение месяца. Не упустите шанс стать легендой!",
    },
  ];

  const toggleQuestion = (index: any) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <SupportContainer>
      <SupportTitle>Поддержка</SupportTitle>
      <FaqList>
        {faqData.map((item, index) => (
          <FaqItem key={index}>
            <QuestionButton onClick={() => toggleQuestion(index)}>
              <QuestionText>{item.question}</QuestionText>
              <ArrowIcon isOpen={openIndex === index}>
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    d="M6 9L12 15L18 9"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </ArrowIcon>
            </QuestionButton>
            <AnswerContainer isOpen={openIndex === index}>
              <AnswerText>{item.answer}</AnswerText>
            </AnswerContainer>
          </FaqItem>
        ))}
      </FaqList>
      <TelegramButton
        style={{ position: "absolute", textTransform: "initial" }}
        href="https://t.me/Onlylazovic"
      >
        Связаться с поддержкой
      </TelegramButton>
    </SupportContainer>
  );
}
