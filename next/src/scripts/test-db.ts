import db from '../lib/db'

const main = async () => {
  try {
    const result = await db.query('SELECT * FROM champions')
    console.log(result)
  } catch (error) { 
    console.error('Error connecting to db', error)
  } finally {
    await db.end()
  }
}

main()
