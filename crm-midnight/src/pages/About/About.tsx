import React from "react";
import {
  AboutContainer,
  AboutSubtitle,
  AboutTitle,
  AboutTitleWrapper,
  ContainerWrapper,
  ImageContainer,
} from "./styles";

export default function About() {
  return (
    <AboutContainer>
      <ImageContainer>
        <AboutTitle>Midnight Club</AboutTitle>
      </ImageContainer>
      <ContainerWrapper>
        <AboutSubtitle>
          MIDNIGHT CLUB — дом спортивного покера! Добро пожаловать
          в эксклюзивное пространство спортивного покера! Обитель для тех, кто
          ценит интеллект, вкус и… немного экстаза! Переступая порог,
          вы окунаетесь в атмосферу уюта: премиальные столы, звон фишек и
          волнующая энергия игры заряжают адреналином!
        </AboutSubtitle>
      </ContainerWrapper>
      <ContainerWrapper>
        <AboutSubtitle>
          <AboutTitleWrapper>Атмосфера</AboutTitleWrapper>
          Здесь играют, заводят полезные знакомства, общаются с интересными
          людьми, отдыхают душой и телом! Это идеальное место, чтобы
          расслабиться после трудового дня и найти то самое состояние
          вдохновения, а может быть и Вашу музу!
        </AboutSubtitle>
      </ContainerWrapper>
      <ContainerWrapper>
        <AboutSubtitle>
          <AboutTitleWrapper>Принципы</AboutTitleWrapper>
          &bull; Спортивный подход — честная игра превыше всего! <br />
          &bull; Уважение к каждому участнику — ценен каждый! <br />
          &bull; Комфортное общение! <br />
          &bull; Эйфория т происходящего — по желанию! <br />
          &bull; Каждый участник подтверждает, что разделяет наши ценности
          и готов к честной борьбе.
        </AboutSubtitle>
      </ContainerWrapper>
      <ContainerWrapper>
        <AboutSubtitle>
          <AboutTitleWrapper>Битва лучших</AboutTitleWrapper>
          Раз в месяц MIDNIGHT собирает самых отважных и умелых игроков сезона.
          27 лучших из лучших сходятся в грандиозном финале! Глубокая игра,
          концентрация и атмосфера великого вечера! Попадание в топ открывает
          двери в систему Golden Ticket - эксклюзивный пропуск в главный финал
          года!
        </AboutSubtitle>
      </ContainerWrapper>
      <ContainerWrapper>
        <AboutSubtitle>
          <AboutTitleWrapper>Обучение</AboutTitleWrapper>
          MIDNIGHT открыт для всех! Даже если вы новичок, мы научим вас играть в
          спортивный покер на специальном бесплатном обучении для новичков!
        </AboutSubtitle>
      </ContainerWrapper>
      <ContainerWrapper>
        <AboutSubtitle>
          MIDNIGHT Club - когда спортивный покер становится изысканным отдыхом,
          поводом для душевного общения и захватывающей игрой с близкими по
          духу.
        </AboutSubtitle>
      </ContainerWrapper>
    </AboutContainer>
  );
}
