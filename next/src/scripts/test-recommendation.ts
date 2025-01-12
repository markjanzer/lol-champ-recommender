import db from '../lib/db'
import { recommendChampions } from '../lib/champions/recommendationEngine'

const main = async () => {
  try {
    const championStats = await db.query('SELECT * FROM champion_stats ORDER BY id DESC LIMIT 1')
    const lastChampionStats = championStats.rows[0].data

    const champSelect = {
      bans: [63],
      allies: [51, 25],
      enemies: [22, 117],
    }

    const result = await recommendChampions(lastChampionStats, champSelect)
    console.log(result)
  } catch (error) { 
    console.error('Error connecting to db', error)
  } finally {
    await db.end()
  }
}

main()