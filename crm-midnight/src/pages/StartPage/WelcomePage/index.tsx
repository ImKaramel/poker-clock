import {
  ArrowButton,
  ArrowContainer,
  BackgroundContainer,
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
} from "./styles";
import background from "../../../assets/startpage_background.svg";
import { ReactComponent as ArrowIcon } from "../../../assets/arrow_forward.svg";
import { ReactComponent as ArrowNext } from "../../../assets/arrow_back_ios.svg";
export default function StartPage() {
  return (
    <StartPageContainer>
      <TitleContainer>
        <Title>Добро пожаловать в Midnight Club</Title>
        <SubTitle>
          Midnight Club — это больше, чем просто спортивный покер. Это место,
          где встречаются единомышленники, где важен не только фарт, но и умение
          читать людей, принимать решения и чувствовать игру.
        </SubTitle>
      </TitleContainer>
      <BackgroundContainer>
        <img src={background} style={{width: '100%', height: '100%', objectFit: 'cover'}}></img>
      </BackgroundContainer>
      <IndicatorContainer>
        <Indicator></Indicator>
        <QuatroIndicator />
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
    </StartPageContainer>
  );
}
