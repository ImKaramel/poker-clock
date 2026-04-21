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
import { ReactComponent as Warning } from "../../assets/warning.svg";
import { GameType } from "../../types";
import { gamesAPI, profileAPI } from "../../utils/api";
import { useParams } from "react-router-dom";
import { InfoChip } from "../Main/styles";
import { getTournamentImage } from "../../utils/tournamentImages";

export default function CurrentTournament() {
  const { id } = useParams<{ id: string }>();

  const [error, setError] = useState("");
  const [game, setGame] = useState<GameType>();
  const [upcomingGames, setUpcomingGames] = useState<number[]>([]);

  const soldOut = 115;

  const participantsCount = game?.participants_count ?? 0;
  const isFull = participantsCount >= soldOut;
  const isRegistered = game ? upcomingGames.includes(game.game_id) : false;

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
    if (!id) return;

    const fetchData = async () => {
      try {
        // 🔥 получаем игру
        const gameResponse = await gamesAPI.getGame(parseInt(id));
        setGame(gameResponse.data);

        // 🔥 получаем профиль (для регистрации)
        const profileResponse = await profileAPI.getProfile();
        const ids = profileResponse.data.upcoming_games.map(
          (g: any) => g.game_id
        );
        setUpcomingGames(ids);
      } catch (err: any) {
        setError(err.message || "Ошибка загрузки");
      }
    };

    fetchData();
  }, [id]);

  const handleRegistry = async () => {
    if (!id) return;

    try {
      if (!isRegistered) {
        await gamesAPI.registerForGame(parseInt(id));
      } else {
        await gamesAPI.discardRegisterForGame(parseInt(id));
      }

      // обновляем профиль после действия
      const response = await profileAPI.getProfile();
      const ids = response.data.upcoming_games.map((g: any) => g.game_id);
      setUpcomingGames(ids);
    } catch (err: any) {
      setError(err.message || "Ошибка регистрации");
    }
  };

  return (
    <CurrentTournamentContainer>
      <TitleContainer>
        <img
          src={getTournamentImage(game?.name || "Загрузка...")}
          style={{ height: "100%", width: "100%", objectFit: "contain" }}
          alt="Выбранный турнир"
        />

        <InfoChip label={formatTime(game?.time || "")} />

        <InfoChip
          label={formatDate(game?.date || "")}
          style={{ top: "151px" }}
        />

        <InfoChip
          label={`Участников: ${participantsCount}`}
          style={{ top: "186px" }}
        />
      </TitleContainer>

      <RulesContainer>
        <RulesWrapper>
          <RulesTitle>Описание</RulesTitle>
          <RulesSubTitle>{game?.name}</RulesSubTitle>

          <RulesTitle>Особенности</RulesTitle>
          <RulesSubTitle style={{ whiteSpace: "pre-line" }}>
            {game?.description}
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

      <JoinButton
        onClick={handleRegistry}
        disabled={isFull && !isRegistered}
        style={{
          opacity: isFull && !isRegistered ? 0.5 : 1,
          cursor: isFull && !isRegistered ? "not-allowed" : "pointer",
        }}
      >
        {isFull && !isRegistered
          ? "Мест нет"
          : isRegistered
          ? "Отменить запись"
          : "Участвовать"}
      </JoinButton>
    </CurrentTournamentContainer>
  );
}