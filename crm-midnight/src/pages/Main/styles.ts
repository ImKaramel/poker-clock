import { Button } from "@mui/material";
import styled from "styled-components";

export const MainContainer = styled.div`
    width: 100%;
    border-radius: 25px;
    background-color: #000000;
`
export const RegisterButton = styled(Button)({
    width: '138px!important',
    height: '35px!important',
    borderRadius: '47px!important',
    padding: '10px 23px!important',
    backgroundColor: '#FEFEFE!important',
    color: '#000!important',
    fontSize: '18px!important',
    fontWeight: '500!important',
})

export const MainAvatarContainer = styled.div`
    width: 55px;
    height: 55px;
    border-radius: 50%;
    overflow: hidden;
`
export const AboutContainer = styled.div`
    width: 100%;
    height: 211px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-radius: 25px;
    margin-top: 5px;
`
export const MainHelpContainer = styled.div`
    position: relative;
    width: 211px;
    height: 211px;
    border-radius: 25px;
    background-color: #14151A;
`
export const MainHelpTitle = styled.div`
    position: absolute;
    top: 23px;
    left: 23px;
    width: auto;
    height: 16px;
    color: #FFFFFF;
    line-height: 83%;
    font-weight: 500;
`
export const MainHelpSubtitle = styled.div`
    height: auto;
    padding: 0 20px 18px 20px;
    font-weight: 400;
    font-size: 13px;
    color: #FFFFFF80;
`