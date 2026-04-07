import { Button } from "@mui/material";
import styled from "styled-components";

export const SupportContainer = styled.div`
    width: 100%;
    border-radius: 25px;
    background-color: #0B0B0B;
`

export const SupportTitle = styled.div`
    position: absolute;
    top: 34px;
    left: 20px;
    width: auto;
    height: 24px;
    font-weight: 500;
    font-size: 24px;
    line-height: 83%;
    color: #FFFFFF;
`

export const FaqList = styled.div`
    position: absolute;
    top: 74px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 100%;
    max-width: 800px;
    margin: 0 auto;
`

export const FaqItem = styled.div`
    width: 100%;
    /* height: 60px; */
    border-radius: 16px;
    background-color: #151A22;
    overflow: hidden;
    transition: all 0.3s ease;
`

export const QuestionButton = styled.button`
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 18px 20px;
    background: transparent;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s ease;
    
    &:hover {
        background-color: #151A22;
    }
`

export const QuestionText = styled.span`
    font-weight: 500;
    font-size: 16px;
    line-height: 83%;
    color: #FFFFFF;
    text-align: left;
    flex: 1;
`

export const ArrowIcon = styled.div<{ isOpen: boolean }>`
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.3s ease;
    transform: ${props => props.isOpen ? 'rotate(180deg)' : 'rotate(0deg)'};
    color: #FFFFFF;
`

export const AnswerContainer = styled.div<{ isOpen: boolean }>`
    max-height: ${props => props.isOpen ? '500px' : '0'};
    opacity: ${props => props.isOpen ? '1' : '0'};
    transition: all 0.3s ease-in-out;
    overflow: hidden;
    border-top: ${props => props.isOpen ? '1px solid #333' : 'none'};
`

export const AnswerText = styled.div`
    padding: 20px;
    font-weight: 300;
    font-size: 15px;
    line-height: 124%;
    color: #FFFFFF;
    background-color: #151A22;
    border-top: 3px solid  #FFFFFF;
    border-radius: 25px;
`
export const TelegramButton = styled(Button)({
    bottom: '100px!important',
    left: '50%!important',
    transform: 'translateX(-50%)!important',
    width: '256px!important',
    height: '44px!important',
    borderRadius: '25px!important',
    padding: '14px 30px!important',
    backgroundColor: '#FFFFFF!important',
    fontWeight: '500!important',
    fontSize: '15px!important',
    color: '#151A22!important'
})