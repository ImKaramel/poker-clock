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
import { Calendar, Check, X } from "lucide-react";
import { ReactComponent as Time } from "../../assets/time.svg";
import butterfly from "../../assets/butterfly_tournament.png";
import { useEffect, useState } from "react";
import { profileAPI, ratingAPI } from "../../utils/api";
import { ProfileType, RatingType } from "../../types";
import { ReactComponent as EditButton } from "../../assets/edit.svg";

const RatingEpxl = 500;

export default function Profile() {
  const [error, setError] = useState("");
  const [profile, setProfile] = useState<ProfileType>();
  const [rating, setRating] = useState<RatingType[]>([]);
  const [visibleRows, setVisibleRows] = useState<RatingType[]>([]);
  const [currentUserId, setCurrentUserId] = useState<string>("");
  const [edited, setEdited] = useState<boolean>(false);
  const [updatedNickname, setUpdatedNickname] = useState<string>("Ваш ник");

  useEffect(() => {
    const getProfile = async () => {
      try {
        const response = await profileAPI.getProfile();
        setProfile(response.data);
      } catch (err: any) {
        setError(err);
      }
    };

    const getRating = async () => {
      try {
        const response = await ratingAPI.getRating();
        setRating(response.data);
      } catch (err: any) {
        setError(err);
      }
    };
    getProfile();
    getRating();
  }, [error]);

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
    try {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const response = await profileAPI.updateProfile(updatedNickname);
      console.log(updatedNickname)
    } catch (err: any) {
      setError(err);
    }
  };

  const calcWidth = () => {
    if (!profile?.user) return 0;
    const width = (profile.user.points / RatingEpxl) * 100;
    return Math.min(width, 100);
  };

  return (
    <ProfileContainer>
      <ProfileInfoContainer>
        <AvatarContainer>
          <img
            src={profile?.user.photo_url}
            style={{ width: "auto" }}
            alt="avatar"
          />
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
                    onClick={() => setEdited(true)}
                    style={{ width: "100%", height: "100%" }}
                    fill="#fff"
                  />
                </EditButtonContainer>
              </div>
            ) : (
              <Wrapper>
                <InputWrapper>
                  <Input
                    value={updatedNickname}
                    onChange={(e) => setUpdatedNickname(e.target.value)}
                  />
                </InputWrapper>

                <Button
                  style={{ color: "#fff" }}
                  onClick={() => setEdited(false)}
                >
                  <X>✕</X>
                </Button>

                <Button
                  style={{ color: "#fff" }}
                  onClick={async () => {
                    await updateNickname();
                    setEdited(false);
                  }}
                >
                  <Check>✓</Check>
                </Button>
              </Wrapper>
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
                Рейтинг {profile?.user?.points || 0}
                <FlashOnIcon sx={{ color: "gold", fontSize: "1rem" }} />
              </RatingSubtitle>
              <RatingSubtitle>
                {RatingEpxl}{" "}
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
            <CurrentTab>Активные</CurrentTab>
            <PastTab>Прошедшие</PastTab>
          </TabContainer>
        </GameHistoryWrapper>
      </GameHistoryContainer>

      <TournamentCardContainer>
        <InfoCardContainer>
          <TournamentName>Butterfly Tournament</TournamentName>
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
            <TimeTitle>5 марта</TimeTitle>
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
            <TimeTitle>19:00</TimeTitle>
          </TimeContainer>
        </InfoCardContainer>
        <img src={butterfly} alt="img" />
      </TournamentCardContainer>
    </ProfileContainer>
  );
}
