// import { faker } from "@faker-js/faker";
import {
  AvatarContainer,
  Button,
  EditButtonContainer,
  GameHistoryContainer,
  GameHistoryTitle,
  GameHistoryWrapper,
  InfoTitle,
  InfoWrapper,
  Input,
  InputWrapper,
  LineContainer,
  Overlay,
  ProfileContainer,
  ProfileInfoContainer,
  ProfileRating,
  ProgressBar,
  RatingSubtitle,
  Wrapper,
} from "./styles";
import FlashOnIcon from "@mui/icons-material/FlashOn";
import RatingTable from "../Rating/RatingTable";
import {
  CurrentTab,
  InfoCardContainer,
  PastTab,
  TabContainer,
  TimeContainer,
  TimeTitle,
  TournamentCardContainer,
  TournamentName,
} from "../Tournaments/styles";
import { Camera, Calendar, Check, X } from "lucide-react";
import { ReactComponent as Time } from "../../assets/time.svg";
import tournament_image from "../../assets/tournament_image.png";
import { ChangeEvent, useEffect, useState } from "react";
import { profileAPI, ratingAPI } from "../../utils/api";
import { ProfileType, RatingType } from "../../types";
import { ReactComponent as EditButton } from "../../assets/edit.svg";

export default function Profile() {
  const [error, setError] = useState("");
  const [profile, setProfile] = useState<ProfileType>();
  const [rating, setRating] = useState<RatingType[]>([]);
  const [visibleRows, setVisibleRows] = useState<RatingType[]>([]);
  const [currentUserId, setCurrentUserId] = useState<string>("");
  const [edited, setEdited] = useState<boolean>(false);
  const [isSavingNickname, setIsSavingNickname] = useState<boolean>(false);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState<boolean>(false);
  const [nick_name, setNick_name] = useState<string>("");
  const [historyTab, setHistoryTab] = useState<"active" | "past">("active");

  useEffect(() => {
    const getProfile = async () => {
      try {
        const response = await profileAPI.getProfile();
        setProfile(response.data);
      } catch (err: any) {
        setError(err?.message || "Не удалось загрузить профиль");
      }
    };

    const getRating = async () => {
      try {
        const response = await ratingAPI.getRating();
        setRating(response.data);
      } catch (err: any) {
        setError(err?.message || "Не удалось загрузить рейтинг");
      }
    };
    getProfile();
    getRating();
  }, []);

  useEffect(() => {
    if (!profile?.user) return;
    setNick_name(profile.user.nick_name || profile.user.first_name || "");
  }, [profile]);

  useEffect(() => {
    if (rating && rating.length > 0 && profile?.user) {
      const currentUserInRating = rating.find(
        (item) => item.user.user_id === profile.user.user_id
      );

      const topRows = rating.slice(0, 6);

      let rows = [...topRows];
      if (
        currentUserInRating &&
        !topRows.some((row) => row.user.user_id === profile.user.user_id)
      ) {
        rows.push(currentUserInRating);
      }

      setVisibleRows(rows);
      setCurrentUserId(profile.user.user_id);
    }
  }, [rating, profile]);

  const updateNickname = async () => {
    const normalizedNickname = nick_name.trim();
    if (!normalizedNickname) {
      setError("Ник не может быть пустым");
      return false;
    }

    try {
      setIsSavingNickname(true);
      setError("");
      const response = await profileAPI.updateProfile(normalizedNickname);

      if (response?.data?.user) {
        setProfile(response.data);
      } else {
        setProfile((prev) =>
          prev
            ? {
                ...prev,
                user: {
                  ...prev.user,
                  nick_name: normalizedNickname,
                },
              }
            : prev
        );
      }

      setNick_name(normalizedNickname);
      return true;
    } catch (err: any) {
      setError(err?.response?.data?.detail || err?.message || "Не удалось сохранить ник");
      return false;
    } finally {
      setIsSavingNickname(false);
    }
  };

  const uploadAvatar = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      setError("Можно загрузить только изображение");
      return;
    }

    try {
      setIsUploadingAvatar(true);
      setError("");
      const response = await profileAPI.uploadAvatar(file);
      setProfile((prev) =>
        prev
          ? {
              ...prev,
              user: response.data.user || {
                ...prev.user,
                photo_url: response.data.photo_url,
              },
            }
          : prev
      );
    } catch (err: any) {
      setError(err?.response?.data?.error || err?.message || "Не удалось загрузить фото");
    } finally {
      setIsUploadingAvatar(false);
      event.target.value = "";
    }
  };

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

  const currentUserInRating = rating.find(
    (item) => item.user.user_id === profile?.user?.user_id
  );
  const currentRatingPoints = currentUserInRating?.points ?? profile?.user?.points ?? 0;
  const qualificationPoints =
    rating[26]?.points ?? rating[rating.length - 1]?.points ?? 0;

  const calcWidth = () => {
    if (qualificationPoints <= 0) return 0;
    const width = (currentRatingPoints / qualificationPoints) * 100;
    return Math.min(width, 100);
  };

  const historyGames =
    historyTab === "active" ? profile?.upcoming_games || [] : profile?.past_games || [];

  return (
    <ProfileContainer>
      <ProfileInfoContainer>
        <AvatarContainer>
          <img
            src={profile?.user.photo_url}
            style={{ width: "auto" }}
            alt="avatar"
          />
          <label style={avatarUploadStyle}>
            <Camera size={18} />
            <input
              type="file"
              accept="image/*"
              onChange={uploadAvatar}
              disabled={isUploadingAvatar}
              style={{ display: "none" }}
            />
          </label>
          <Overlay />
          <InfoWrapper>
            {!edited ? (
              <div
                style={{
                  display: "flex",
                  alignItems: "baseline",
                  justifyContent: "flex-start",
                }}
              >
                <InfoTitle style={{ width: "auto" }}>
                  {profile?.user?.nick_name || profile?.user?.first_name}
                </InfoTitle>
                <EditButtonContainer>
                  <EditButton
                    onClick={() => {
                      setNick_name(profile?.user?.nick_name || profile?.user?.first_name || "");
                      setEdited(true);
                    }}
                    style={{ width: "100%", height: "100%" }}
                    fill="#fff"
                  />
                </EditButtonContainer>
              </div>
            ) : (
              <Wrapper>
                <InputWrapper>
                  <Input
                    value={nick_name}
                    onChange={(e) => setNick_name(e.target.value)}
                  />
                </InputWrapper>

                <Button
                  style={{ color: "#fff" }}
                  onClick={() => {
                    setNick_name(profile?.user?.nick_name || profile?.user?.first_name || "");
                    setEdited(false);
                  }}
                >
                  <X>✕</X>
                </Button>

                <Button
                  style={{ color: "#fff" }}
                  onClick={async () => {
                    if (isSavingNickname) return;
                    const isSaved = await updateNickname();
                    if (isSaved) {
                      setEdited(false);
                    }
                  }}
                  disabled={isSavingNickname}
                >
                  <Check>{isSavingNickname ? "…" : "✓"}</Check>
                </Button>
              </Wrapper>
            )}
            {!!error && (
              <div style={{ fontSize: 12, color: "#ff8585", marginTop: 6 }}>{error}</div>
            )}

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
              <RatingSubtitle>
                Рейтинг {currentRatingPoints}
                <FlashOnIcon sx={{ color: "gold", fontSize: "1rem" }} />
              </RatingSubtitle>
              <RatingSubtitle>
                {qualificationPoints}
                <FlashOnIcon sx={{ color: "gold", fontSize: "1rem" }} />
              </RatingSubtitle>
            </div>
          </InfoWrapper>
        </AvatarContainer>
      </ProfileInfoContainer>

      <ProfileRating>
        <RatingTable rows={visibleRows} currentUserId={currentUserId} />
      </ProfileRating>

      <GameHistoryContainer>
        <GameHistoryWrapper>
          <GameHistoryTitle>История игр</GameHistoryTitle>
          <TabContainer style={{ position: "inherit" }}>
            <CurrentTab
              onClick={() => setHistoryTab("active")}
              style={historyTab === "active" ? activeTabStyle : inactiveTabStyle}
            >
              Активные
            </CurrentTab>
            <PastTab
              onClick={() => setHistoryTab("past")}
              style={historyTab === "past" ? activeTabStyle : inactiveTabStyle}
            >
              Прошедшие
            </PastTab>
          </TabContainer>
        </GameHistoryWrapper>
      </GameHistoryContainer>
      {historyGames.map((item, index) => (
        <TournamentCardContainer key={index}>
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
          <img src={tournament_image} alt="img" />
        </TournamentCardContainer>
      ))}
    </ProfileContainer>
  );
}

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

const avatarUploadStyle = {
  position: "absolute" as const,
  right: 18,
  top: 18,
  width: 38,
  height: 38,
  borderRadius: "50%",
  background: "rgba(0,0,0,0.58)",
  color: "#fff",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  border: "1px solid rgba(255,255,255,0.28)",
  cursor: "pointer",
  zIndex: 3,
};
