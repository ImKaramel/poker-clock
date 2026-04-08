export interface ProfileType {
  upcoming_games: Array<any>,
  user: {
    created_at: string,
    first_name: string,
    is_admin: boolean,
    is_banned: boolean,
    points: number,
    total_games_played: number,
    user_id: string,
    username: string
  }
}