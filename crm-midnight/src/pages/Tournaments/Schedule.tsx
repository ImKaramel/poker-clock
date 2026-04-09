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
import butterfly from "../../assets/butterfly_tournament.png";
import { GameType } from "../../types";
import { gamesAPI } from "../../utils/api";
import { useNavigate } from "react-router-dom";

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
            <TournamentName>{item.description}</TournamentName>
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
              <TimeTitle>{item.date}</TimeTitle>
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
              <TimeTitle>{item.time}</TimeTitle>
            </TimeContainer>
          </InfoCardContainer>
          <img src={butterfly} alt="img" />
        </TournamentCardContainer>
      ))}
    </TournamentsContainer>
  );
}
