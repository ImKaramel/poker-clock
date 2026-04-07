import React from "react";
import {
  CurrentTournamentContainer,
  JoinButton,
  RulesContainer,
  RulesSubTitle,
  RulesTitle,
  RulesWrapper,
  TitleContainer,
  WarningContainer,
  WarningSubtitle,
  WarningTitle,
  WarningWrapper,
} from "./styles";
import current_tournament from "../../assets/current_tournament.jpg";
import { ReactComponent as Warning } from "../../assets/warning.svg";

export default function CurrentTournament() {
  return (
    <CurrentTournamentContainer>
      <TitleContainer>
        <img
          src={current_tournament}
          style={{ height: "100%", width: "100%", objectFit: "contain" }}
        />
      </TitleContainer>
      <RulesContainer>
        <RulesWrapper>
          <RulesTitle>Общие правила</RulesTitle>
          <RulesSubTitle>
            Техасский холдем — классический формат без упрощений.
          </RulesSubTitle>
          <RulesTitle>Особенности</RulesTitle>
          <RulesSubTitle>
            &bull; Re-entry ограничен: до 3 входов на игрока <br />
            &bull; Без бонусных фишек и утилит <br />
            &bull; Гарантия рейтинговых очков: 20 000
          </RulesSubTitle>
          <WarningContainer>
            <WarningWrapper>
              <div
                style={{
                  width: "197px",
                  height: "22px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                }}
              >
                <Warning />
                <WarningTitle>Правила отмены</WarningTitle>
              </div>
              <WarningSubtitle>
                  Пожалуйста, отменяйте регистрацию заранее, если не планируете
                  приходить, чтобы не занимать место у участников из очереди.
                </WarningSubtitle>
            </WarningWrapper>
          </WarningContainer>
        </RulesWrapper>
      </RulesContainer>
      <JoinButton>
        Участвовать
      </JoinButton>
    </CurrentTournamentContainer>
  );
}
