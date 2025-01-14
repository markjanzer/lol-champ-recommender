'use client';
import { useEffect, useState } from "react";
import { recommendChampions } from "@/lib/champions/recommendationEngine";
import { ChampionDataMap, ChampionPerformance } from "@/lib/types/champions"

interface Props {
  championStats: ChampionDataMap;
  championIds: number[];
}

export default function ChampionRecommender({championStats, championIds}: Props) {
  const [allies, setAllies] = useState<number[]>([0,0,0,0]);
  const [enemies, setEnemies] = useState<number[]>([0,0,0,0,0]);
  const [recommendations, setRecommendations] = useState<ChampionPerformance[]>([]);

  const handleAllyChange = (index: number, value: number) => {
    const newAllies = [...allies];
    newAllies[index] = value;
    setAllies(newAllies);
  };

  const handleEnemyChange = (index: number, value: number) => {
    const newEnemies = [...enemies];
    newEnemies[index] = value;
    setEnemies(newEnemies);
  };

  useEffect(() => {
    console.log('Effect triggered with:', {
      allies,
      enemies,
      championStats: Object.keys(championStats).length
    });

    const validAllies = allies.filter(ally => championIds.includes(ally));
    const validEnemies = enemies.filter(enemy => championIds.includes(enemy));

    if (validAllies.length === 0 && validEnemies.length === 0) {
      setRecommendations([]);
      return;
    }
    
    const recommendations = recommendChampions(championStats, { allies: validAllies, enemies: validEnemies, bans: [] });
    setRecommendations(recommendations);
  }, [championStats, championIds, allies, enemies]);


  return (
    <>
      <div>
        <h1 className="text-2xl font-bold">Pick Champions</h1>
        <div className="mt-4">
          <h2 className="text-lg font-bold">Allies</h2>
          {allies.map((ally, index) => (
            <input
              key={index}
              className="border border-gray-300 rounded-md p-2 text-black"
              type="number"
              min={1}
              max={10000}
              value={ally || ''}
              onChange={(e) => handleAllyChange(index, parseInt(e.target.value) || 0)}
            />
          ))}
        </div>
        <div className="mt-4">
          <h2 className="text-lg font-bold">Enemies</h2>
          {enemies.map((enemy, index) => (
            <input
              key={index}
              className="border border-gray-300 rounded-md p-2 text-black"
              type="number"
              min={1}
              max={10000}
              value={enemy || ''}
              onChange={(e) => handleEnemyChange(index, parseInt(e.target.value) || 0)}
            />
          ))}
        </div>
      </div>
      <div className="mt-4">
        <h2 className="text-2xl font-bold">Selected Champions</h2>
        <div className="mt-4">
          {recommendations.map((recommendation) => (
            <div key={recommendation.championId}>
              {recommendation.championId}
            </div>
          ))}
        </div>
      </div>
    </>
  )
}