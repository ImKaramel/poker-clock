import React, { useEffect, useState } from "react";
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
import current_tournament from "../../assets/grand_opening.jpg";
import { ReactComponent as Warning } from "../../assets/warning.svg";
import { GameType } from "../../types";
import { gamesAPI } from "../../utils/api";
import { useParams } from "react-router-dom";
import { InfoChip } from "../Main/styles";

export default function CurrentTournament() {
  const { id } = useParams<{ id: string }>();
  const [error, setError] = useState("");
  const [game, setGame] = useState<GameType>();
  const [registered, setRegistered] = useState<boolean>(false);

  const formatTime = (timeStr: string) => {
    if (timeStr && timeStr.includes(":")) {
      return timeStr.split(":").slice(0, 2).join(":");
    }
    return timeStr || "Время не указано";
  };

  const formatDate = (dateStr: string) => {
    const months = [
      "января",
      "февраля",
      "марта",
      "апреля",
      "мая",
      "июня",
      "июля",
      "августа",
      "сентября",
      "октября",
      "ноября",
      "декабря",
    ];

    const date = new Date(dateStr);
    const day = date.getDate();
    const month = months[date.getMonth()];

    return `${day} ${month}`;
  };

  useEffect(() => {
    const getGames = async () => {
      if (!id) return;
      try {
        const response = await gamesAPI.getGame(parseInt(id));
        setGame(response.data);
      } catch (err: any) {
        setError(err);
      }
    };
    getGames();
  }, [id, error]);

  const Registry = async () => {
    if (!id) return;
    if (!registered) {
      try {
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const response = await gamesAPI.registerForGame(parseInt(id));
        setRegistered(true);
      } catch (err: any) {
        setError(err);
      }
    } else {
      try {
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const response = await gamesAPI.discardRegisterForGame(parseInt(id));
        setRegistered(false);
      } catch (err: any) {
        setError(err);
      }
    }
  };
  return (
    <CurrentTournamentContainer>
      <TitleContainer>
        <img
          src={current_tournament}
          style={{ height: "100%", width: "100%", objectFit: "contain" }}
          alt="Выбранный турнир"
        />
        <InfoChip
          label={formatTime(game?.time || game?.time || "")}
          style={{
            justifyContent: "flex-start",
          }}
        />
        <InfoChip
          label={formatDate(game?.date || game?.time || "")}
          style={{
            fontWeight: "500!important",
            top: "151px",
            width: "193px",
            justifyContent: "flex-start",
          }}
        />
      </TitleContainer>
      <RulesContainer>
        <RulesWrapper>
          <RulesTitle>Описание</RulesTitle>
          <RulesSubTitle>{game?.name}</RulesSubTitle>
          <RulesTitle>Особенности</RulesTitle>
          <RulesSubTitle>
            {game?.description}
            &bull; Гарантия рейтинговых очков: {game?.base_points}
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
      <JoinButton onClick={Registry}>
        {registered ? "Вы зарегестрированы" : "Участвовать"}
      </JoinButton>
    </CurrentTournamentContainer>
  );
}
