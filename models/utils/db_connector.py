from sqlalchemy import create_engine
import pandas as pd
import os
from dotenv import load_dotenv

load_dotenv()

def get_db_connection():
  db_url = os.getenv("DATABASE_URL")
  return create_engine(db_url)

def get_matchup_data():
  engine = get_db_connection()
  query = """
  SELECT * FROM matchups
  """
  return pd.read_sql(query, engine)

def get_champion_stats():
  engine = get_db_connection()
  query = """
  SELECT * FROM champion_stats ORDER BY created_at DESC LIMIT 1
  """
  return pd.read_sql(query, engine)

def get_first_match():
  engine = get_db_connection()
  query = """
  SELECT * FROM matches ORDER BY created_at DESC LIMIT 1
  """
  return pd.read_sql(query, engine)