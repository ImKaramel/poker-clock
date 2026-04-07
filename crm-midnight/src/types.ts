export interface User {
    user_id: number;
    username: string;
    first_name: string;
    name: string
    last_name: string;
    photo?: string;
    email: string;
    date_of_birth: string;
    phone: string;
    status: boolean;
    is_admin: boolean;
    is_banned: boolean;
  }
  export interface Tournament {
    description: string;
    date: string;
    location: string;
    participants: string;
    status: string;
    photo_id: string
  }
  export interface Player {
    rank: number;
    name: string;
    score: number;
    isYou?: boolean;
  }

  export interface Me {
    id: number;
    username: string;
    first_name: string;
    last_name: string;
    role: string;
  }