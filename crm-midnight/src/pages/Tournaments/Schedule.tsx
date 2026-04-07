import React from "react";
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
import butterfly from '../../assets/butterfly_tournament.png'

export default function Schedule() {
  return (
    <TournamentsContainer>
      <HeaderContainer>
        <TitleTournaments>Турниры</TitleTournaments>
        <TabContainer>
          <CurrentTab>Текущие</CurrentTab>
          <PastTab>Прошедшие</PastTab>
        </TabContainer>
      </HeaderContainer>
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
    </TournamentsContainer>
  );
}
