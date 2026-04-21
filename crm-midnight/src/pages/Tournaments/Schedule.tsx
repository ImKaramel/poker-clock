import React, { useEffect, useState } from "react";
import {
  CurrentTab,
  HeaderContainer,
  InfoCardContainer,
  PastTab,
  TabContainer,
  TimeContainer,
  TimeTitle,
  TitleTournaments,
  TournamentCardContainer,
  TournamentName,
  TournamentsContainer,
} from "./styles";
import { ReactComponent as Calendar } from "../../assets/calendar_date.svg";
import { ReactComponent as Time } from "../../assets/time.svg";
// import tournament_image from "../../assets/tournament_image.png";
import { GameType } from "../../types";
import { gamesAPI } from "../../utils/api";
import { useNavigate } from "react-router-dom";
import DEEPSTACK from '../../assets/DEEP STACK.png'
import FREEZOUT from '../../assets/FREEZOUT.png'
import MIDNIGHT_GIFT from '../../assets/MIDNIGHT Gift.png'
import MIDNIGHT_Phoenix from '../../assets/MIDNIGHT Phoenix.png'

const games_img = [
  MIDNIGHT_Phoenix,
  MIDNIGHT_GIFT,
  DEEPSTACK,
  FREEZOUT
]

export default function Schedule() {
  const [error, setError] = useState("");
  const [games, setGames] = useState<GameType[]>();
  useEffect(() => {
    const getGames = async () => {
      try {
        const response = await gamesAPI.getGames();
        setGames(response.data);
      } catch (err: any) {
        setError(err);
      }
    };
    getGames();
  }, [error]);
  const formatTime = (timeStr?: string) => {
    if (!timeStr) return "Время не указано";
    if (timeStr.includes(":")) {
      return timeStr.split(":").slice(0, 2).join(":");
    }
    return timeStr;
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return "Дата не указана";
    
    const months = [
      "января", "февраля", "марта", "апреля", "мая", "июня",
      "июля", "августа", "сентября", "октября", "ноября", "декабря",
    ];

    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return dateStr;
    
    const day = date.getDate();
    const month = months[date.getMonth()];
    return `${day} ${month}`;
  };

  const navigate = useNavigate();

  return (
    <TournamentsContainer>
      <HeaderContainer>
        <TitleTournaments>Турниры</TitleTournaments>
        <TabContainer>
          <CurrentTab>Текущие</CurrentTab>
          <PastTab>Прошедшие</PastTab>
        </TabContainer>
      </HeaderContainer>
      {games?.map((item, index) => (
        <TournamentCardContainer
          key={index}
          onClick={() => {
            navigate(`/games/${item.game_id}`);
          }}
          
        >
          <InfoCardContainer>
            <TournamentName>{item.name}</TournamentName>
            <TimeContainer>
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  display: "flex",
                  alignItems: "center",
                }}
              >
                <Calendar />
              </div>
              <TimeTitle>{formatDate(item.date)}</TimeTitle>
            </TimeContainer>
            <TimeContainer style={{ gridColumn: "2/3" }}>
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  display: "flex",
                  alignItems: "center",
                }}
              >
                <Time />
              </div>
              <TimeTitle>{formatTime(item.time)}</TimeTitle>
            </TimeContainer>
          </InfoCardContainer>
          <img src={games_img[index]} alt="img" />
        </TournamentCardContainer>
      ))}
    </TournamentsContainer>
  );
}
