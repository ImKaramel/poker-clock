import styled from "styled-components";
import about_background from '../../assets/About_background.svg'

export const AboutContainer =  styled.div`
    width: 100%;
    border-radius: 25px;
    background-color: #0B0B0B;
`
export const ImageContainer = styled.div`
    position: relative;
    width: 100%;
    height: 388px;
    background-image: url(${about_background});
    background-repeat: no-repeat;
    background-position: center;
`
export const AboutTitle = styled.div`
    position: absolute;
    top: 321px;
    left: 20px;
    width: auto;
    height: 45px;
    font-weight: 500;
    font-size: 45px;
    color: #FFFFFF;
    line-height: 83%;
`
export const ContainerWrapper = styled.div`
    width: 100%;
    /* height: 204px; */
    border-radius: 25px;
    background-color: #151A22;
    margin-top: 5px;
`
export const AboutTitleWrapper = styled.div`
    height: 24px;
    width: auto;
    font-weight: 500;
    font-size: 24px;
    line-height: 83%;
    color: #FFFFFF;
    margin-bottom: 20px;
`
export const AboutSubtitle = styled.div`
    padding: 20px 30px;
    font-weight: 400;
    font-size: 16px;
    line-height: 131%;
    color: #FFFFFF;
    text-align: left;
`