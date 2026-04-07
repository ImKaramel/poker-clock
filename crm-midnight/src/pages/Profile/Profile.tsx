import { faker } from "@faker-js/faker";
import {
  AvatarContainer,
  GameHistoryContainer,
  GameHistoryTitle,
  GameHistoryWrapper,
  InfoTitle,
  InfoWrapper,
  LineContainer,
  Overlay,
  ProfileContainer,
  ProfileInfoContainer,
  ProfileRating,
  ProgressBar,
  RatingSubtitle,
} from "./styles";
import FlashOnIcon from "@mui/icons-material/FlashOn";
import RatingTable, { rows } from "../Rating/RatingTable";
import { CurrentTab, InfoCardContainer, PastTab, TabContainer, TimeContainer, TimeTitle, TournamentCardContainer, TournamentName } from "../Tournaments/styles";
import { Calendar } from "lucide-react";
import { ReactComponent as Time } from "../../assets/time.svg";
import butterfly from '../../assets/butterfly_tournament.png'

const currentUser = rows[10];
const visibleRows = rows.slice(0, 6);
if (!visibleRows.some((r) => r.id === currentUser.id)) {
  visibleRows.push(currentUser);
}

const RatingEpxl = 500;

const generateAvatar = () => {
  return {
    nickname: faker.person.firstName(),
    avatar: faker.image.avatar(),
    rating: faker.number.int({ min: 50, max: 500 }),
  };
};
export const User = generateAvatar();

export const calcWidth = () => {
  return (User.rating / RatingEpxl) * 100;
};

export default function Profile() {
  return (
    <ProfileContainer>
      <ProfileInfoContainer>
        <AvatarContainer>
          <img src={User.avatar} style={{ width: "auto" }} />
          <Overlay />
          <InfoWrapper>
            <InfoTitle>{User.nickname}</InfoTitle>
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
                Рейтинг {User.rating}
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
        <RatingTable rows={visibleRows} currentUserId={currentUser.id} />
      </ProfileRating>
      <GameHistoryContainer>
        <GameHistoryWrapper>
          <GameHistoryTitle>История игр</GameHistoryTitle>
          <TabContainer style={{position: 'inherit'}}>
            <CurrentTab>Активные</CurrentTab>
            <PastTab>Прошедшие</PastTab>
          </TabContainer>
        </GameHistoryWrapper>
      </GameHistoryContainer>
      <TournamentCardContainer>
        <InfoCardContainer>
          <TournamentName>Butterfly Tournament</TournamentName>
          <TimeContainer>
            <div style={{ width: "12px", height: "12px", display: 'flex', alignItems: 'center' }}>
              <Calendar />
            </div>
            <TimeTitle>5 марта</TimeTitle>
          </TimeContainer>
          <TimeContainer style={{gridColumn: '2/3'}}>
            <div style={{ width: "12px", height: "12px", display: 'flex', alignItems: 'center' }}>
              <Time />
            </div>
            <TimeTitle>19:00</TimeTitle>
          </TimeContainer>
        </InfoCardContainer>
        <img src={butterfly}/>
      </TournamentCardContainer>
    </ProfileContainer>
  );
}
