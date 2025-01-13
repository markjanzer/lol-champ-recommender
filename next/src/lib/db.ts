import { Pool } from 'pg'
// import dotenv from 'dotenv'
// import path from 'path'

// dotenv.config({ path: path.resolve(__dirname, '../../../.env') })

const pool = new Pool({
  user: process.env.POSTGRES_USER,
  password: process.env.POSTGRES_PASSWORD,
  host: process.env.POSTGRES_HOST,
  database: process.env.POSTGRES_DATABASE,
  port: parseInt(process.env.POSTGRES_PORT || '5432'),
})

pool.query('SELECT NOW()', (err) => {
  if (err) {
    console.error('Database connection test failed:', err)
  } else {
    console.log('Database connected successfully')
  }
})

if (process.env.POSTGRES_DATABASE !== 'lol_champ_recommender_development') {
  console.error('Database connection test failed: Database is set to', process.env.POSTGRES_DATABASE)
}

export default pool