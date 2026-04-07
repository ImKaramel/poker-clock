import React from "react";
import {
  AboutContainer,
  MainAvatarContainer,
  MainContainer,
  MainHelpContainer,
  MainHelpSubtitle,
  MainHelpTitle,
  RegisterButton,
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
import { calcWidth, User } from "../Profile/Profile";
import RatingTable, { rows } from "../Rating/RatingTable";
import { ReactComponent as LogoVector } from "../../assets/logo_vector.svg";
import { useNavigate } from "react-router";

const RatingEpxl = 500;
const currentUser = rows[10];
const visibleRows = rows.slice(0, 3);
if (!visibleRows.some((r) => r.id === currentUser.id)) {
  visibleRows.push(currentUser);
}

export default function Main() {
  const navigate = useNavigate();
  return (
    <MainContainer>
      <TitleContainer>
        <img
          src={current_tournament}
          style={{ height: "100%", width: "100%", objectFit: "contain" }}
        />
        <RegisterButton
          style={{
            textTransform: "none",
            position: "absolute",
            top: "225px",
            left: "30px",
          }}
        >
          Записаться
        </RegisterButton>
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
          />
        </MainAvatarContainer>
        <InfoWrapper
          style={{
            position: "inherit",
            height: "55px",
            padding: 0,
            width: "calc(100%  - 70px)",
          }}
        >
          <InfoTitle style={{ fontSize: "16px" }}>{User.nickname}</InfoTitle>
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
              Рейтинг {User.rating}
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
        <RatingTable rows={visibleRows} currentUserId={currentUser.id} />
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
        <MainHelpContainer style={{ display: "flex", alignItems: "end" }} onClick={() => navigate("/support")}>
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
