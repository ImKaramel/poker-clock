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
import tournament_image from "../../assets/tournament_image.png";
import { GameType, TournamentHistoryType } from "../../types";
import { gamesAPI, tournamentHistoryAPI } from "../../utils/api";
import { useNavigate } from "react-router-dom";

export default function Schedule() {
  const [error, setError] = useState("");
  const [games, setGames] = useState<GameType[]>();
  const [history, setHistory] = useState<TournamentHistoryType[]>([]);
  const [tab, setTab] = useState<"current" | "past">("current");

  useEffect(() => {
    const getGames = async () => {
      try {
        const response = await gamesAPI.getGames();
        setGames(response.data);
      } catch (err: any) {
        setError(err);
      }
    };
    const getHistory = async () => {
      try {
        const response = await tournamentHistoryAPI.getHistory();
        setHistory(response.data);
      } catch (err: any) {
        setError(err?.message || "Не удалось загрузить историю");
      }
    };
    getGames();
    getHistory();
  }, []);
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
  const weekAgo = new Date();
  weekAgo.setDate(weekAgo.getDate() - 7);
  weekAgo.setHours(0, 0, 0, 0);

  const recentHistory = history.filter((item) => {
    const date = new Date(item.date);
    return !isNaN(date.getTime()) && date >= weekAgo;
  });

  return (
    <TournamentsContainer>
      <HeaderContainer>
        <TitleTournaments>Турниры</TitleTournaments>
        <TabContainer>
          <CurrentTab
            onClick={() => setTab("current")}
            style={tab === "current" ? activeTabStyle : inactiveTabStyle}
          >
            Текущие
          </CurrentTab>
          <PastTab
            onClick={() => setTab("past")}
            style={tab === "past" ? activeTabStyle : inactiveTabStyle}
          >
            Прошедшие
          </PastTab>
        </TabContainer>
      </HeaderContainer>
      {tab === "current" &&
        games?.map((item, index) => (
          <TournamentCardContainer
            key={index}
            onClick={() => {
              navigate(`/games/${item.game_id}`);
            }}
          >
            <InfoCardContainer>
              <TournamentName>{item.name}</TournamentName>
              <TimeContainer>
                <div style={iconBoxStyle}>
                  <Calendar />
                </div>
                <TimeTitle>{formatDate(item.date)}</TimeTitle>
              </TimeContainer>
              <TimeContainer style={{ gridColumn: "2/3" }}>
                <div style={iconBoxStyle}>
                  <Time />
                </div>
                <TimeTitle>{formatTime(item.time)}</TimeTitle>
              </TimeContainer>
            </InfoCardContainer>
            <img src={tournament_image} alt="img" />
          </TournamentCardContainer>
        ))}

      {tab === "past" &&
        recentHistory.map((item) => (
          <TournamentCardContainer key={item.id} style={{ height: "auto", minHeight: 92 }}>
            <InfoCardContainer style={{ width: "100%", height: "auto", paddingRight: 8 }}>
              <TournamentName>{item.tournament_name}</TournamentName>
              <TimeContainer>
                <div style={iconBoxStyle}>
                  <Calendar />
                </div>
                <TimeTitle>{formatDate(item.date)}</TimeTitle>
              </TimeContainer>
              <TimeContainer style={{ gridColumn: "2/3" }}>
                <div style={iconBoxStyle}>
                  <Time />
                </div>
                <TimeTitle>{formatTime(item.time || undefined)}</TimeTitle>
              </TimeContainer>
              <div style={resultStyle}>
                {item.participants_count} игроков
                {item.participants?.length
                  ? ` · ${item.participants
                      .slice()
                      .sort((a, b) => (a.position || 999) - (b.position || 999))
                      .slice(0, 3)
                      .map((p) => `${p.position ? `${p.position}. ` : ""}${p.first_name || p.username}${p.final_points ? ` (${p.final_points})` : ""}`)
                      .join(", ")}`
                  : ""}
              </div>
            </InfoCardContainer>
            <img src={tournament_image} alt="img" />
          </TournamentCardContainer>
        ))}
      {!!error && <div style={{ color: "#ff8585", padding: 16 }}>{String(error)}</div>}
    </TournamentsContainer>
  );
}

const iconBoxStyle = {
  width: "12px",
  height: "12px",
  display: "flex",
  alignItems: "center",
};

const activeTabStyle = {
  backgroundColor: "#fff",
  border: "1px solid #fff",
  color: "#151a22",
  cursor: "pointer",
};

const inactiveTabStyle = {
  backgroundColor: "transparent",
  border: "1px solid #7e7e7e",
  color: "#ffffff60",
  cursor: "pointer",
};

const resultStyle = {
  gridColumn: "1 / 3",
  color: "#ffffff99",
  fontSize: 12,
  lineHeight: "16px",
};
