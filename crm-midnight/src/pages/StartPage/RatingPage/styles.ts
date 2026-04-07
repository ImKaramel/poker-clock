import styled from "styled-components";
import { StartPageContainer } from "../WelcomePage/styles";
import rating_background from "../../../assets/rating_background.png";

export const BackgroundContainer = styled.div`
  /* height: calc(100dvh - 47px); */
  min-height: 590px;
  flex: 1;
  background: 
    linear-gradient(
      to bottom,
      rgba(0, 0, 0, 0) 60%,
      rgba(0, 0, 0, 0.8) 100%
    ),
    url(${rating_background});

  background-repeat: no-repeat;
  background-position: top;
  background-size: cover;
`;

export const InfoContainer = styled.div`
  position: absolute;
  top: 397px;
  left: 50%;
  transform: translateX(-50%);
  width: 388px;
  height: 185px;
  border-radius: 25px;
  border: 1px solid transparent;

  background: linear-gradient(#151a22, #151a2299) padding-box,
    linear-gradient(180deg, #f6f6f6 0%, #94949438 22%) border-box;
  backdrop-filter: blur(2.6px);
`;
export const InfoInner = styled.div`
  width: 100%;
  height: 100%;
  background: #151a2299;
  border-radius: 24px;
  display: flex;
  justify-content: space-evenly;
`;

export const PointsWrapper = styled.div`
  width: 101px;
  height: 82px;
  position: relative;
  top: -4px;
`;
export const PointsIndicator = styled.div`
  position: absolute;
  bottom: -2px;
  left: 50%;
  transform: translateX(-50%);
  width: 44px;
  height: 36px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
`;
export const UserText = styled.div`
  width: 44px;
  height: 12px;
  font-weight: 500;
  font-size: 14px;
  line-height: 83%;
  color: #fff;
`
export const UserNumber = styled.div`
  display: flex;
  align-items: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: linear-gradient(to right, #F6F6F6, #949494);
`
export const UserNumberText = styled.div`
  width: 7px;
  height: 9px;
  margin: 0 auto;
  color: #151A22;
  font-size: 10.4px;
  font-weight: 500;
  line-height: 83%;
  text-align: center;
`
export const UserImage = styled.div`
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  z-index: 999;
  width: 40px;
  height: 40px;
  border: 1px solid #ffffff57;
  border-radius: 50%;
  overflow: hidden;
`;
export const UserBackground = styled.div`
  position: absolute;
  top: 24px;
  width: 100%;
  height: 51px;
  border-radius: 10px;
  background: linear-gradient(45deg, #F6F6F6, #94949438);
  padding: 1px
`;

export const UserBackgroundInner = styled.div`
  background:  rgb(27, 30, 35);
  border-radius: 10px;
  height: 100%;
  backdrop-filter: blur(32.6px);
`;

export const PointsContainer = styled.div`
  position: absolute;
  bottom: -25px;
  left: 50%;
  transform: translateX(-50%);
  width: 46px;
  height: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
`
export const PointsText = styled.div`
  width: 32px;
  height: 12px;
  font-weight: 500;
  font-size: 12px;
  color: #fff;
  text-align: center;
`
export const RatingExplContainer = styled.div`
  width: 100%;
  height: 67px;
  background: rgba(255, 255, 255, 0.1);
  position: absolute;
  bottom: 0;
  border-radius: 25px;
  display: flex;
  align-items: center;
  justify-content: space-between;
`
export const RatingUserInfo = styled.div`
  width: 160px;
  height: 27px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-left: 17px;
`
export const Points = styled.div`
  width: 16px;
  height: 11px;
  font-weight: 500;
  font-size: 13px;
  line-height: 83%;
  color: #fff;
`
export const AccountCircle = styled.div`
  width: 27px;
  height: 27px;
  display: flex;
  justify-content: center;
  align-items: center;
`
export const UsernameContainer = styled.div`
  width: 99px;
  height: 12px;
  font-weight: 500;
  font-size: 12px;
  color: #fff;
`
export const PointsContainerExmpl = styled.div`
  width: 39px;
  height: 12px;
  margin-right: 17px;
  display: flex;
  align-items: center;
`
export const BottomContainer = styled.div`
    margin-bottom: 47px;
    display: flex;
    flex-direction: column;
    align-items: center;
`