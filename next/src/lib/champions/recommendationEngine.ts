import { ChampionPerformance, ChampionInteraction, ChampionDataMap, ChampSelect } from '../types/champions';

export async function recommendChampions(
  championStats: ChampionDataMap,
  champSelect: ChampSelect
): Promise<ChampionPerformance[]> {
  const allChampIds = Object.keys(championStats).map(Number)
  const unavailableChampIds = [...champSelect.bans, ...champSelect.allies, ...champSelect.enemies]

  const results: ChampionPerformance[] = []

  for (const champId of allChampIds) {
    if (unavailableChampIds.includes(champId)) {
      continue
    }

    const championPerformance: ChampionPerformance = {
      championId: champId,
      winProbability: -1,
      synergies: [],
      matchups: [],
    }

    for (const allyId of champSelect.allies) {
      const synergy = championStats[champId].synergies[allyId]
      if (!synergy) {
        console.error(`Synergy not found for champion ${champId} and ally ${allyId}`)
        continue
      }

      let winProbability: number = 0.50
      if (synergy.games > 0) {
        // Smoothing winrate
        winProbability = (synergy.wins + 5) / (synergy.games + 10)
      }

      const interaction: ChampionInteraction = {
        championId: allyId,
        winProbability: winProbability,
        wins: synergy.wins,
        games: synergy.games,
      }

      championPerformance.synergies.push(interaction)
    }

    for (const enemyId of champSelect.enemies) {
      const matchup = championStats[champId].matchups[enemyId]
      if (!matchup) {
        console.error(`Matchup not found for champion ${champId} and enemy ${enemyId}`)
        continue
      }

      let winProbability: number = 0.50
      if (matchup.games > 0) {
        winProbability = (matchup.wins + 5) / (matchup.games + 10)
      }

      const interaction: ChampionInteraction = {
        championId: enemyId,
        winProbability: winProbability,
        wins: matchup.wins,
        games: matchup.games,
      }

      championPerformance.matchups.push(interaction)
    }

    let winProbability: number = 0.0
    const dataPoints = championPerformance.synergies.length + championPerformance.matchups.length
    if (dataPoints > 0) {
      for (const synergy of championPerformance.synergies) {
        winProbability += synergy.winProbability
      }

      for (const matchup of championPerformance.matchups) {
        winProbability += matchup.winProbability
      }

      winProbability /= dataPoints
    } else {
      winProbability = 0.50
    }

    championPerformance.winProbability = winProbability

    results.push(championPerformance)
  }

  results.sort((a, b) => b.winProbability - a.winProbability)
  
  return results;
}