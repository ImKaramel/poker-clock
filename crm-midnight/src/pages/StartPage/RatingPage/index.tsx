import React from "react";
import {
  ArrowButton,
  ArrowContainer,
  Indicator,
  IndicatorContainer,
  NavigationContainer,
  NextArrow,
  NextArrowsContainer,
  QuatroIndicator,
  RectangleText,
  RoundedRectangleContainer,
  StartPageContainer,
  SubTitle,
  Title,
  TitleContainer,
} from "../WelcomePage/styles";
import {
  AccountCircle,
  BackgroundContainer,
  BottomContainer,
  InfoContainer,
  InfoInner,
  Points,
  PointsContainer,
  PointsContainerExmpl,
  PointsIndicator,
  PointsText,
  PointsWrapper,
  RatingExplContainer,
  RatingUserInfo,
  UserBackground,
  UserBackgroundInner,
  UserImage,
  UsernameContainer,
  UserNumber,
  UserNumberText,
  UserText,
} from "./styles";
import UserImage_1 from "../../../assets/userImage_1.jpg";
import UserImage_2 from "../../../assets/userImage_2.png";
import UserImage_3 from "../../../assets/userImage_3.jpg";
import { ReactComponent as Lightning } from "../../../assets/lightning.svg";
import { ReactComponent as Lightning_Gold } from "../../../assets/lightning_gold.svg";
import { ReactComponent as Account } from "../../../assets/account.svg";
import { ReactComponent as ArrowIcon } from "../../../assets/arrow_forward.svg";
import { ReactComponent as ArrowNext } from "../../../assets/arrow_back_ios.svg";

export default function RatingPage() {
  return (
    <StartPageContainer style={{height:'100vh'}}>
      <TitleContainer
        style={{ paddingLeft: 0, paddingRight: 0, width: "100%", gap: 0, height: '100%'}}
      >
        <Title style={{ padding: "0 20px 0 20px" }}>Система рейтинга</Title>
        <BackgroundContainer>
          <SubTitle style={{ padding: "0 20px 0 20px" }}>
            Умная система рейтинга, которая считает ваши победы, нокауты
            и начисляет очки. Сражайтесь за место в лидерборде!
          </SubTitle>
          <InfoContainer>
            <InfoInner>
              <PointsWrapper>
                <UserImage>
                  <img
                    src={UserImage_1}
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover",
                    }}
                  />
                </UserImage>
                <UserBackground>
                  <UserBackgroundInner></UserBackgroundInner>
                </UserBackground>
                <PointsIndicator>
                  <UserText>Brilliant</UserText>
                  <UserNumber>
                    <UserNumberText>2</UserNumberText>
                  </UserNumber>
                </PointsIndicator>
                <PointsContainer>
                  <PointsText>1948</PointsText>
                  <Lightning />
                </PointsContainer>
              </PointsWrapper>
              <PointsWrapper
                style={{ width: "123px", height: "82px", top: "-26px" }}
              >
                <UserImage
                  style={{ backgroundColor: "rgb(112, 112, 112", top: "0" }}
                >
                  <img
                    src={UserImage_2}
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover",
                    }}
                  />
                </UserImage>
                <UserBackground style={{ height: "58px", top: "15px" }}>
                  <UserBackgroundInner></UserBackgroundInner>
                </UserBackground>
                <PointsIndicator
                  style={{ width: "49px", height: "36px", bottom: "0" }}
                >
                  <UserText style={{ width: "49px", height: "12px" }}>
                    BesObr
                  </UserText>
                  <UserNumber
                    style={{
                      background: "linear-gradient(to right, #F8E84E, #E4D43A)",
                    }}
                  >
                    <UserNumberText>1</UserNumberText>
                  </UserNumber>
                </PointsIndicator>
                <PointsContainer style={{ width: "75px", height: "19px" }}>
                  <PointsText
                    style={{ fontSize: "19px", width: "53px", height: "19px" }}
                  >
                    1948
                  </PointsText>
                  <Lightning_Gold />
                </PointsContainer>
              </PointsWrapper>
              <PointsWrapper>
                <UserImage>
                  <img
                    src={UserImage_3}
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover",
                    }}
                  />
                </UserImage>
                <UserBackground>
                  <UserBackgroundInner></UserBackgroundInner>
                </UserBackground>
                <PointsIndicator>
                  <UserText>PokerPro</UserText>
                  <UserNumber
                    style={{
                      background: "linear-gradient(to right, #F2CBA9, #D88B53)",
                    }}
                  >
                    <UserNumberText>3</UserNumberText>
                  </UserNumber>
                </PointsIndicator>
                <PointsContainer>
                  <PointsText>900</PointsText>
                  <Lightning />
                </PointsContainer>
              </PointsWrapper>
            </InfoInner>
            <RatingExplContainer>
              <RatingUserInfo>
                <Points>35</Points>
                <AccountCircle>
                  <Account />
                </AccountCircle>
                <UsernameContainer>Username (Вы)</UsernameContainer>
              </RatingUserInfo>
              <PointsContainerExmpl>
                <PointsText>900</PointsText>
                <Lightning />
              </PointsContainerExmpl>
            </RatingExplContainer>
          </InfoContainer>
        </BackgroundContainer>
      </TitleContainer>
      <BottomContainer>
        <IndicatorContainer>
          <QuatroIndicator />
          <Indicator></Indicator>
          <QuatroIndicator />
          <QuatroIndicator />
        </IndicatorContainer>
        <NavigationContainer>
          <ArrowContainer>
            <ArrowButton>
              <ArrowIcon width={16} height={16} color="white"></ArrowIcon>
            </ArrowButton>
          </ArrowContainer>
          <RoundedRectangleContainer>
            <ArrowContainer style={{ backgroundColor: "#fff" }}>
              <ArrowButton>
                <ArrowIcon
                  width={16}
                  height={16}
                  color="black"
                  transform="rotate(180)"
                ></ArrowIcon>
              </ArrowButton>
            </ArrowContainer>
            <RectangleText>Старт</RectangleText>
            <NextArrowsContainer>
              <NextArrow>
                <ArrowNext opacity="8%" />
              </NextArrow>
              <NextArrow>
                <ArrowNext opacity="24%" />
              </NextArrow>
              <NextArrow>
                <ArrowNext />
              </NextArrow>
            </NextArrowsContainer>
          </RoundedRectangleContainer>
        </NavigationContainer>
      </BottomContainer>
    </StartPageContainer>
  );
}
