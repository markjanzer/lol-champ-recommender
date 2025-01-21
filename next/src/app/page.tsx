import ChampionRecommender from "@/components/ChampionRecommender";
import champions from '@/data/champions.json'
import championStats from '@/data/champion_stats.json'

export default async function Home() {
  return (
    <div className="grid mx-auto max-w-6xl mt-10">
      <div className="col-span-3">
        <ChampionRecommender championStats={championStats} champions={champions} />
      </div>
    </div>
  );
}