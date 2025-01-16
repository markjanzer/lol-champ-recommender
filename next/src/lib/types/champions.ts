export interface Champion {
  name: string;
  api_id: number;
}

export interface ChampionPerformance {
  championId: number;
  winProbability: number;
  synergies: ChampionInteraction[];
  matchups: ChampionInteraction[];
}

export interface ChampionInteraction {
  championId: number;
  winProbability: number;
  wins: number;
  games: number;
}

export interface ChampSelect {
  bans: number[];
  allies: number[];
  enemies: number[];
}

export interface WinStats {
  wins: number;
  games: number;
}

export interface ChampionData {
  winrate: WinStats;
  matchups: Record<number, WinStats>;
  synergies: Record<number, WinStats>;
}

export type ChampionDataMap = Record<number, ChampionData>;