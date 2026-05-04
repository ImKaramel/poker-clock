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
  const [selectedHistory, setSelectedHistory] = useState<TournamentHistoryType | null>(null);

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
          <TournamentCardContainer
            key={item.id}
            style={pastCardStyle}
            onClick={() => setSelectedHistory(item)}
          >
            <div style={pastInfoStyle}>
              <TournamentName style={pastOnlyNameStyle}>{item.tournament_name}</TournamentName>
            </div>
            <img src={tournament_image} alt="img" style={pastImageStyle} />
          </TournamentCardContainer>
        ))}
      {!!error && <div style={{ color: "#ff8585", padding: 16 }}>{String(error)}</div>}

      {selectedHistory && (
        <div style={overlayStyle} onClick={() => setSelectedHistory(null)}>
          <div style={modalStyle} onClick={(event) => event.stopPropagation()}>
            <div style={modalHeaderStyle}>
              <div style={modalTitleStyle}>{selectedHistory.tournament_name}</div>
              <button type="button" style={closeButtonStyle} onClick={() => setSelectedHistory(null)}>
                x
              </button>
            </div>

            <div style={playersListStyle}>
              {selectedHistory.participants
                .slice()
                .sort((a, b) => {
                  const aPoints = a.final_points || 0;
                  const bPoints = b.final_points || 0;
                  return bPoints - aPoints;
                })
                .map((player, index) => (
                  <div key={player.id} style={playerRowStyle}>
                    <div style={playerNameStyle}>
                      {index + 1}. {player.first_name || player.username}
                    </div>
                    <div style={playerPointsStyle}>{player.final_points || 0}</div>
                  </div>
                ))}
            </div>
          </div>
        </div>
      )}
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

const pastCardStyle = {
  minHeight: 88,
  height: 88,
  alignItems: "center",
  justifyContent: "space-between",
  overflow: "hidden",
  boxSizing: "border-box" as const,
  cursor: "pointer",
};

const pastInfoStyle = {
  minWidth: 0,
  flex: "1 1 auto",
  padding: "18px 12px 18px 20px",
  display: "flex",
  flexDirection: "column" as const,
  justifyContent: "center",
};

const pastOnlyNameStyle = {
  height: "auto",
  minHeight: 0,
  lineHeight: "20px",
  display: "-webkit-box",
  WebkitLineClamp: 2,
  WebkitBoxOrient: "vertical" as const,
  overflow: "hidden",
};

const pastImageStyle = {
  width: 112,
  minWidth: 112,
  height: "100%",
  objectFit: "cover" as const,
  opacity: 0.72,
};

const overlayStyle = {
  position: "fixed" as const,
  inset: 0,
  background: "rgba(0, 0, 0, 0.72)",
  zIndex: 30,
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  padding: 20,
};

const modalStyle = {
  width: "100%",
  maxWidth: 420,
  maxHeight: "80vh",
  overflow: "auto",
  background: "#151A22",
  borderRadius: 24,
  padding: 20,
  boxSizing: "border-box" as const,
};

const modalHeaderStyle = {
  display: "flex",
  alignItems: "flex-start",
  justifyContent: "space-between",
  gap: 12,
  marginBottom: 18,
};

const modalTitleStyle = {
  color: "#fff",
  fontSize: 22,
  lineHeight: "28px",
  fontWeight: 600,
};

const closeButtonStyle = {
  border: "none",
  background: "transparent",
  color: "#ffffff99",
  fontSize: 24,
  lineHeight: "24px",
  cursor: "pointer",
  padding: 0,
};

const playersListStyle = {
  display: "flex",
  flexDirection: "column" as const,
  gap: 10,
};

const playerRowStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: 12,
  padding: "12px 14px",
  borderRadius: 14,
  background: "rgba(255,255,255,0.04)",
};

const playerNameStyle = {
  color: "#fff",
  fontSize: 15,
  lineHeight: "20px",
  minWidth: 0,
};

const playerPointsStyle = {
  color: "#fff",
  fontSize: 15,
  lineHeight: "20px",
  fontWeight: 700,
  flexShrink: 0,
};
