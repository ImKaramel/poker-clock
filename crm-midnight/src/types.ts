export interface ProfileType {
  upcoming_games: GameType[],
  past_games: GameType[],
  user: {
    created_at: string,
    first_name: string,
    is_admin: boolean,
    is_banned: boolean,
    nick_name?: string
    points: number,
    total_games_played: number,
    user_id: string,
    username: string,
    photo_url: string,
  }
}

export interface GameType {
  name: string,
  base_points: number,
  buyin: number,
  date: string,
  description: string,
  game_id: number,
  is_active: boolean,
  location: string,
  min_players_for_extra_points: number,
  participants_count: number,
  participants_details: Array<any>,
  photo: any,
  points_per_extra_player: number,
  reentry_buyin: number,
  time: string
}

export interface TournamentHistoryParticipant {
  id: number,
  user_id: string,
  username: string,
  first_name: string,
  last_name: string,
  entries: number,
  rebuys: number,
  addons: number,
  total_spent: number,
  payment_method?: string | null,
  payment_method_display?: string,
  position?: number | null,
  final_points?: number,
}

export interface TournamentHistoryType {
  id: number,
  game: number,
  date: string,
  time?: string | null,
  tournament_name: string,
  location: string,
  buyin: number,
  reentry_buyin?: number | null,
  total_revenue: number,
  participants_count: number,
  completed_at: string,
  participants: TournamentHistoryParticipant[],
}

export interface RatingType {
  games_played: number,
  points: number,
  rank: number,
  user: {
    created_at: string,
    first_name: string,
    is_admin: boolean,
    is_banned: boolean,
    points: number,
    total_games_played: number,
    user_id: string,
    username: string
    photo_url: string
    nick_name?: string
  }
}
export interface GameParticipant {
  addons: number,
  entries: number,
  final_points: number,
  game: string,
  id: number,
  joined_at: string,
  position: any,
  rebuys: number,
  user: string,
  user_info: {
    first_name: string,
    last_name: string,
    user_id: number,
    username: string
  }
}
