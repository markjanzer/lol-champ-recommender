import { recommendChampions } from '../lib/champions/recommendationEngine'
import championStats from '@/data/champion_stats.json'

const main = async () => {
  try {

    const lastChampionStats = championStats

    const champSelect = {
      bans: [63],
      allies: [3, 67],
      enemies: [22, 117],
    }

    const result = recommendChampions(lastChampionStats, champSelect)
    console.log(result)
  } catch (error) { 
    console.error('Error connecting to db', error)
  } finally {
    await db.end()
  }
}

main()