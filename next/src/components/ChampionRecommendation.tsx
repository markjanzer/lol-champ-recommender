import { ChampionPerformance, Champion } from "@/lib/types/champions"

interface Props {
  championPerformance: ChampionPerformance;
  champions: Champion[];
}

interface ChampionStatTableProps {
  title: string;
  stats: Array<{championId: number; winProbability: number}>;
  getChampionName: (id: number) => string | undefined;
}

function getColor(winrate: number) {
  let intensity = 0;
  
  if (winrate > 0.5) {
    intensity = Math.min((winrate - 0.5) * 10, 1);
  } else {
    intensity = Math.min((0.5 - winrate) * 10, 1);
  }

  return winrate > 0.5 
    ? `rgba(59, 130, 246, ${0.2 + intensity * 0.8})` // blue
    : `rgba(249, 115, 22, ${0.2 + intensity * 0.8})`; // orange
}

function formatWinrate(winrate: number) {
  return (winrate * 100).toFixed(2) + "%";
}

function renderWinrate(winrate: number) {
  return (
    <span 
      className="underline" 
      style={{ textDecorationColor: getColor(winrate) }}
    >
      {formatWinrate(winrate)}
    </span>
  );
}

function ChampionStatTable({ title, stats, getChampionName }: ChampionStatTableProps) {
  if (stats.length === 0) {
    return null;
  }
  return (
    <div className="flex flex-col">
      <p className="text-md font-bold text-center">{title}</p>
      <table>
        <tbody>
          {stats.map(stat => (
            <tr key={stat.championId}>
              <td className="pr-4">{getChampionName(stat.championId)}</td>
              <td className="text-right">{formatWinrate(stat.winProbability)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default function ChampionRecommendation({championPerformance, champions}: Props) {
  const getChampionName = (apiId: number) => {
    return champions.find(champion => champion.api_id === apiId)?.name;
  };

  return (
    <div key={championPerformance.championId} className="mt-4 border-b pb-4">
      <p className="text-lg font-bold">{getChampionName(championPerformance.championId)} {renderWinrate(championPerformance.winProbability)}</p>
      <div className="flex flex-row justify-around mt-2">
        <ChampionStatTable
          title="Synergies" 
          stats={championPerformance.synergies} 
          getChampionName={getChampionName} 
        />
        <ChampionStatTable 
          title="Matchups" 
          stats={championPerformance.matchups} 
          getChampionName={getChampionName} 
        />
      </div>
    </div>
  )
}