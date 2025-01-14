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
    <div className="grid grid-cols-4 mx-auto max-w-6xl mt-10">
      <div className="col-span-1">
        <h2 className="text-2xl font-bold">Champions</h2>
        <table className="table-fixed">
          <tbody>
            {champions.map((champion) => (
              <tr key={champion.api_id}>
                <td>{champion.name}</td>
                <td>{champion.api_id}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Probably don't want to keep this fixed, but it helps for now. */}
      <div className="col-span-3 fixed top-10 right-40">
        <ChampionRecommender championStats={championStats} championIds={champions.map(champion => champion.api_id)} />
      </div>
    </div>
  );
}