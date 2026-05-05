import midnight_phoenix from "../assets/MIDNIGHT Phoenix.jpg";
import deep_stack from "../assets/DEEP STACK.jpg";
import freezout from "../assets/FREEZOUT.jpg";
import midnight_gift from "../assets/MIDNIGHT Gift.jpg";
import freeroll from "../assets/FREEROLL.jpg";
import speed_racer from "../assets/SPEED RACER.jpg";
import classic_holdem from "../assets/HOLDEM Classic.jpg";
import midnight_poker from "../assets/MIDNIGHT Poker.jpg";
import midnight_knokout from "../assets/MIDNIGHT KNOKOUT.jpg";
import contract_tournament from "../assets/CONTRACT TOURNAMENT.jpg";
import defaultTournamentImage from "../assets/tournament_image.png";

const TOURNAMENT_IMAGES_MAP: Record<string, string> = {
  "MIDNIGHT PHOENIX": midnight_phoenix,
  "DEEP STACK": deep_stack,
  "FREEZOUT": freezout,
  "FREEZEOUT": freezout,
  "MIDNIGHT GIFT": midnight_gift,
  "FREEROLL": freeroll,
  "SPEED RACER": speed_racer,
  "CLASSIC HOLDEM": classic_holdem,
  "MIDNIGHT POKER": midnight_poker,
  "MIDNIGHT KNOKOUT": midnight_knokout,
  "CONTRACT TOURNAMENT": contract_tournament,
};

const normalizeTournamentName = (tournamentName: string): string =>
  tournamentName
    .trim()
    .replace(/\s+/g, " ")
    .toUpperCase();

export const getTournamentImage = (tournamentName: string): string => {
  if (!tournamentName) {
    return defaultTournamentImage;
  }

  const normalizedName = normalizeTournamentName(tournamentName);
  const directMatch = TOURNAMENT_IMAGES_MAP[normalizedName];

  if (directMatch) {
    return directMatch;
  }

  console.warn(`Картинка для турнира "${tournamentName}" не найдена, используется дефолтная`);
  return defaultTournamentImage;
};

export const hasCustomTournamentImage = (tournamentName: string): boolean => {
  return normalizeTournamentName(tournamentName) in TOURNAMENT_IMAGES_MAP;
};
