import React, { useEffect, useState } from "react";
import {
  AboutContainer,
  InfoChip,
  MainAvatarContainer,
  MainContainer,
  MainHelpContainer,
  MainHelpSubtitle,
  MainHelpTitle,
} from "./styles";
import { TitleContainer } from "../Tournaments/styles";
import current_tournament from "../../assets/current_tournament.jpg";
import {
  InfoTitle,
  InfoWrapper,
  LineContainer,
  ProfileInfoContainer,
  ProfileRating,
  ProgressBar,
  RatingSubtitle,
} from "../Profile/styles";
import FlashOnIcon from "@mui/icons-material/FlashOn";
import RatingTable from "../Rating/RatingTable";
import { ReactComponent as LogoVector } from "../../assets/logo_vector.svg";
import { useNavigate } from "react-router-dom";
import { gamesAPI, profileAPI, ratingAPI } from "../../utils/api";
import { GameType, ProfileType, RatingType } from "../../types";
import { User } from "../Profile/Profile";

const RatingEpxl = 500;

export default function Main() {
  const [profile, setProfile] = useState<ProfileType | null>(null);
  const [rating, setRating] = useState<RatingType[]>([]);
  const [nearestGame, setNearestGame] = useState<GameType | null>(null);

  useEffect(() => {
    const loadProfile = async () => {
      try {
        const response = await profileAPI.getProfile();
        setProfile(response.data);
      } catch (err: any) {
        console.error("Error fetching profile:", err);
      }
    };

    const loadRating = async () => {
      try {
        const response = await ratingAPI.getRating();
        setRating(response.data);
      } catch (err: any) {
        console.error("Error fetching rating:", err);
      } 
    };

    const loadGames = async () => {
      try {
        const response = await gamesAPI.getGames();
        const allGames: GameType[] = response.data;
        if (allGames && allGames.length > 0) {
          const nearest = findNearestGame(allGames);
          setNearestGame(nearest);
        }
      } catch (err: any) {
        console.error("Error fetching games:", err);
      } 
    };

    loadProfile();
    loadRating();
    loadGames();
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

  const findNearestGame = (gamesList: GameType[]) => {
    const now = new Date();

    const parseGameDate = (game: GameType) => {
      return new Date(`${game.date}T${game.time}`);
    };

    const upcomingGames = gamesList.filter((game) => {
      const gameDate = parseGameDate(game);
      return gameDate >= now;
    });

    if (upcomingGames.length === 0) {
      const sorted = [...gamesList].sort((a, b) => {
        return parseGameDate(a).getTime() - parseGameDate(b).getTime();
      });
      return sorted[0] || null;
    }

    const sortedUpcoming = upcomingGames.sort((a, b) => {
      return parseGameDate(a).getTime() - parseGameDate(b).getTime();
    });

    return sortedUpcoming[0];
  };

  const calcWidth = () => {
    if (!profile?.user) return 0;
    const width = (profile.user.points / RatingEpxl) * 100;
    return Math.min(width, 100);
  };

  const currentUserInRating = rating?.find(
    (item) => item.user.user_id === profile?.user?.user_id
  );

  const currentUserId = currentUserInRating?.user?.user_id || profile?.user?.user_id;

  const getVisibleRows = (): RatingType[] => {
    if (!rating || rating.length === 0) return [];
    
    const topRows = rating.slice(0, 3);
    
    if (currentUserInRating && topRows.some(row => row.user.user_id === currentUserId)) {
      return topRows;
    }
    
    if (currentUserInRating && !topRows.includes(currentUserInRating)) {
      return [...topRows, currentUserInRating];
    }
    
    return topRows;
  };

  const visibleRows = getVisibleRows();
  const navigate = useNavigate();

  const handleTournamentClick = () => {
    if (nearestGame) {
      navigate(`/games/${nearestGame.game_id}`);
      console.log('ближайшая игра', nearestGame)
    }
  };



  return (
    <MainContainer>
      <TitleContainer onClick={handleTournamentClick}>
        <img
          src={current_tournament}
          style={{ height: "100%", width: "100%", objectFit: "contain" }}
          alt="current_tournament"
        />
        {nearestGame && (
          <>
            <InfoChip
              label={formatTime(nearestGame.time)}
              style={{
                justifyContent: "flex-start",
              }}
            />
            <InfoChip
              label={formatDate(nearestGame.date)}
              style={{
                fontWeight: "500!important",
                top: "151px",
                width: "193px",
                justifyContent: "flex-start",
              }}
            />
          </>
        )}
      </TitleContainer>

      <ProfileInfoContainer
        style={{
          height: "95px",
          backgroundColor: "#14151A",
          marginTop: "5px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          width: "calc(100% - 40px)",
          margin: "0 auto",
          gap: "14px",
          padding: "0 20px",
        }}
      >
        <MainAvatarContainer>
          <img
            src={User.avatar}
            style={{ width: "100%", height: "100%", objectFit: "cover" }}
            alt="avatar"
          />
        </MainAvatarContainer>
        <InfoWrapper
          style={{
            position: "inherit",
            height: "55px",
            padding: 0,
            width: "calc(100% - 70px)",
          }}
        >
          <InfoTitle style={{ fontSize: "16px" }}>
            {profile?.user?.first_name || "Игрок"}
          </InfoTitle>
          <LineContainer>
            <ProgressBar style={{ width: `${calcWidth()}%` }} />
          </LineContainer>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <RatingSubtitle style={{ fontSize: "12px" }}>
              Рейтинг {profile?.user?.points || 0}
              <FlashOnIcon sx={{ color: "gold", fontSize: "1rem" }} />
            </RatingSubtitle>
            <RatingSubtitle style={{ fontSize: "12px" }}>
              {RatingEpxl}{" "}
              <FlashOnIcon sx={{ color: "gold", fontSize: "1rem" }} />
            </RatingSubtitle>
          </div>
        </InfoWrapper>
      </ProfileInfoContainer>

      <ProfileRating>
        <RatingTable 
          rows={visibleRows} 
          currentUserId={currentUserId || ""}
        />
      </ProfileRating>

      <AboutContainer>
        <MainHelpContainer
          style={{
            backgroundColor: "#4C4D52",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
          onClick={() => navigate("/about")}
        >
          <MainHelpTitle>О клубе</MainHelpTitle>
          <LogoVector />
        </MainHelpContainer>
        <MainHelpContainer
          style={{ display: "flex", alignItems: "end" }}
          onClick={() => navigate("/support")}
        >
          <MainHelpTitle>Поддержка</MainHelpTitle>
          <MainHelpSubtitle>
            Поможем с записью, оплатой и любыми вопросами по клубу
          </MainHelpSubtitle>
        </MainHelpContainer>
      </AboutContainer>

      <ProfileInfoContainer
        style={{
          position: "relative",
          height: "150px",
          display: "flex",
          alignItems: "end",
        }}
      >
        <MainHelpTitle>Адрес</MainHelpTitle>
        <MainHelpSubtitle>Москва, ул. Новослободская 39</MainHelpSubtitle>
      </ProfileInfoContainer>
    </MainContainer>
  );
}