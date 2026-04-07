import styled from "styled-components";

export const TournamentsContainer = styled.div`
  width: 100%;
  height: 100%;
  background-color: #0b0b0b;
  border-radius: 25px;
`;
export const HeaderContainer = styled.div`
  width: 100%;
  height: 132px;
  background-color: rgb(17, 20, 25);
  border-radius: 25px;
`;
export const TitleTournaments = styled.div`
  position: absolute;
  top: 34px;
  left: 20px;
  width: 94px;
  height: 20px;
  font-weight: 500;
  font-size: 24px;
  line-height: 83%;
  color: #fff;
`;
export const TabContainer = styled.div`
  position: absolute;
  top: 74px;
  left: 20px;
  width: 361px;
  height: 38px;
  display: flex;
  align-items: center;
`;

export const CurrentTab = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 119px;
  height: 38px;
  border-radius: 25px;
  /* padding: 14px 30px 14px 30px; */
  background-color: #fff;
  font-size: 15px;
  font-weight: 400;
  color: #151a22;
`;
export const PastTab = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 142px;
  height: 38px;
  border-radius: 25px;
  border: 1px solid #7e7e7e;
  /* padding: 14px 30px 14px 30px; */
  background-color: transparent;
  font-size: 15px;
  font-weight: 400;
  color: #ffffff60;
`;
export const TournamentCardContainer = styled.div`
  width: 100%;
  height: 79px;
  border-radius: 25px;
  background: linear-gradient(to right, rgb(17, 20, 25), rgb(30, 32, 36));
  margin-top: 5px;
  display: flex;
  align-items: center;
`;

export const InfoCardContainer = styled.div`
  width: 164px;
  height: 39px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: auto auto;
  gap: 12px;
  padding: 20px;
`;

export const TournamentName = styled.div`
  grid-column: 1 / 3;
  font-size: 15px;
  font-weight: 500;
  color: #fff;
  height: 15px;
`;
export const TimeContainer = styled.div`
  grid-column: 1 / 2;
  width: 64px;
  height: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 5px;
`;
export const TimeTitle = styled.div`
  width: 47px;
  height: 12px;
  font-weight: 400;
  font-size: 12px;
  color: #fff;
`;

export const CurrentTournamentContainer = styled.div`
  width: 100%;
  border-radius: 25px;
  background-color: #0B0B0B;
`
export const TitleContainer = styled.div`
  width: 100%;
  height: 388px;
  border-radius: 25px;
`
export const RulesContainer = styled.div`
  width: 100%;
  height: 553px;
  border-radius: 25px;
  background-color: #151A22;
`
export const RulesWrapper  = styled.div`
  height: 100%;
  padding: 34px 20px;
`
export const RulesTitle = styled.div`
  width: auto;
  height: 24px;
  font-weight: 500;
  font-size: 24px;
  line-height: 83%;
  color: #fff;
  margin-bottom: 19px;
`
export const RulesSubTitle = styled.div`
  width: auto;
  height: auto;
  font-weight: 400;
  font-size: 16px;
  line-height: 131%;
  color: #fff;
  margin-bottom: 25px;
`

export const WarningContainer = styled.div`
  width: 100%;
  height: 200px;
  border-radius: 25px;
  background-color: rgb(11,11,11);
  border-top: 5px solid #FFFFFF;
`

export const WarningWrapper = styled.div`
  padding: 30px 32px;
`

export const WarningTitle = styled.div`
  width: 160px;
  height: 20px;
  font-weight: 500;
  font-size: 20px;
  line-height: 83%;
  color: #FFFFFF;
`

export const WarningSubtitle = styled.div`
  width: 100%;
  height: 100px;
  font-weight: 300;
  font-size: 15px;
  line-height: 131%;
  color: #FFFFFF;
  margin-top: 17px;
`
export const JoinButton = styled.div`
  width: 100%;
  height: 85px;
  border-radius: 25px;
  background-color: #FEFEFE;
  font-size: 22px;
  font-weight: 500;
  color: #000000;
  display: flex;
  position: fixed;
  bottom: 0;
  align-items: center;
  justify-content: center;
`