import db from "@/lib/db";
import ChampionRecommender from "@/components/ChampionRecommender";

const fetchChampions = async () => {
  const result = await db.query('SELECT name, api_id FROM champions ORDER BY name ASC');
  return result.rows;
};

const fetchChampionStats = async () => {
  const result = await db.query('SELECT * FROM champion_stats ORDER BY id DESC LIMIT 1');
  return result.rows[0].data;
};

export default async function Home() {
  const champions = await fetchChampions();
  const championStats = await fetchChampionStats();

  return (
    <div className="grid mx-auto max-w-6xl mt-10">
      <div className="col-span-3">
        <ChampionRecommender championStats={championStats} champions={champions} />
      </div>
    </div>
  );
}