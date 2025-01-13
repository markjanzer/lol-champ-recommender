import db from '../lib/db'

const main = async () => {
  try {
    const championStats = await db.query('SELECT * FROM champion_stats ORDER BY id DESC LIMIT 1')

    const lastChampionStats = championStats.rows[0].data
    console.log(lastChampionStats['1'])
  } catch (error) { 
    console.error('Error connecting to db', error)
  } finally {
    await db.end()
  }
}

main()
