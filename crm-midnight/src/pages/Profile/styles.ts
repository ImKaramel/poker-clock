import styled from "styled-components";

export const ProfileContainer = styled.div`
    width: 100%;
    border-radius: 25px;
    background-color: #0B0B0B;
    min-height: 100vh;
`
export const ProfileInfoContainer = styled.div`
    width: 100%;
    height: 388px;
    border-radius: 25px;
`
export const AvatarContainer = styled.div`
    position: relative;
    overflow: hidden;
    height: 388px;
    object-fit: contain;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border-radius: 25px;
`
export const Overlay = styled.div`
    position: absolute;
    inset: 0;
    background: linear-gradient(to bottom, rgba(0,0,0,0) 40%, rgba(0,0,0,0.9) 100%);
`
export const InfoWrapper = styled.div`
    position: absolute;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    top: 293px;
    left: 0;
    width: calc(100% - 32px);
    height: 80px;
    padding: 0 16px 28px 16px;
`
export const InfoTitle = styled.div`
    width: 100%;
    height: 36px;
    font-weight: 500;
    font-size: 36px;
    color: #fff;
`
export const LineContainer = styled.div`
    width: 100%;
    height: 8px;
    border-radius: 72px;
    background-color: #49474882;

`
export const ProgressBar = styled.div`
    height: 100%;
    background: linear-gradient(to right, #232326, #FEFEFE);
    border-radius: 72px;
`
export const RatingSubtitle = styled.div`
    display: flex;
    align-items: center;
    width: auto;
    height: 15px;
    font-weight: 500;
    font-size: 15px;
    color: #FFFFFF;
`
export const ProfileRating = styled.div`
    width: calc(100% - 54px);
    height: auto;
    border-radius: 25px;
    background-color: #151A22;
    margin-top: 5px;
    padding: 27px;
`
export const GameHistoryContainer = styled.div`
    width: 100%;
    height: 108px;
    border-radius: 25px;
    background-color: #151A22;
`
export const GameHistoryWrapper = styled.div`
    width: auto;
    height: 70px;
    padding: 22px 0 20px 20px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
`
export const GameHistoryTitle = styled.div`
    width: auto;
    height: 16px;
    font-weight: 500;
    line-height: 83%;
    color: #FFFFFF;
`