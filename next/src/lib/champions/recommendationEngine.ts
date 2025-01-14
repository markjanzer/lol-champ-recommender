import { ChampionPerformance, WinStats, ChampionInteraction, ChampionDataMap, ChampSelect } from '../types/champions';

export function recommendChampions(
  championStats: ChampionDataMap,
  champSelect: ChampSelect
): ChampionPerformance[] {
  const allChampIds = Object.keys(championStats).map(Number)
  const unavailableChampIds = [...champSelect.bans, ...champSelect.allies, ...champSelect.enemies]
  const availableChampIds = allChampIds.filter(id => !unavailableChampIds.includes(id))

  return availableChampIds
    .map(champId => getChampionPerformance(champId, championStats, champSelect))
    .sort((a, b) => b.winProbability - a.winProbability)
}

function getChampionPerformance(champId: number, championStats: ChampionDataMap, champSelect: ChampSelect): ChampionPerformance {
  const synergies = champSelect.allies.map(allyId => createInteraction(championStats[champId].synergies[allyId], allyId))
  const matchups = champSelect.enemies.map(enemyId => createInteraction(championStats[champId].matchups[enemyId], enemyId)) 
  const winProbability = calculateWinProbability([...synergies, ...matchups])

  return {
    championId: champId,
    winProbability: winProbability,
    synergies: synergies,
    matchups: matchups,
  }
}

function createInteraction(stats: WinStats, championId: number): ChampionInteraction {
  if (stats === null) {
    console.error(`Stats not found for champion ${championId}`)
    return {
      championId: championId,
      winProbability: 0.50,
      wins: 0,
      games: 0,
    }
  }

  let winProbability: number = 0.50
  console.log(stats)
  if (stats.games > 0) {
    // Smoothing winrate
    winProbability = (stats.wins + 5) / (stats.games + 10)
  }

  return {
    championId: championId,
    winProbability: winProbability,
    wins: stats.wins,
    games: stats.games,
  }
}

function calculateWinProbability(interactions: ChampionInteraction[]): number {
  if (interactions.length === 0) {
    return 0.50
  }

  return interactions.reduce((sum, interaction) => sum + interaction.winProbability, 0) / interactions.length
}