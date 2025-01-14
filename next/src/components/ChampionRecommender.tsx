'use client';
import { useEffect, useState } from "react";
import { recommendChampions } from "@/lib/champions/recommendationEngine";
import { ChampionDataMap, ChampionPerformance } from "@/lib/types/champions"

interface Props {
  championStats: ChampionDataMap;
}

export default function ChampionRecommender({championStats}: Props) {
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

    const inputAllies = allies.filter(ally => ally !== 0);
    const inputEnemies = enemies.filter(enemy => enemy !== 0);

    if (inputAllies.length === 0 && inputEnemies.length === 0) {
      return;
    }
    
    const recommendations = recommendChampions(championStats, { allies: inputAllies, enemies: inputEnemies, bans: [] });
    setRecommendations(recommendations);
  }, [championStats, allies, enemies]);


  return (
    <>
      <div>
        <h1 className="text-2xl font-bold">Pick Champions</h1>
        <div className="mt-4">
          <h2 className="text-lg font-bold">Allies</h2>
          {allies.map((ally, index) => (
            <input
              key={index}
              className="border border-gray-300 rounded-md p-2"
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
              className="border border-gray-300 rounded-md p-2"
              type="number"
              min={1}
              max={10000}
              value={enemy || ''}
              onChange={(e) => handleEnemyChange(index, parseInt(e.target.value) || 0)}
            />
          ))}
        </div>
        <div className="mt-4">
          <button className="bg-blue-500 text-white rounded-md p-2">Recommend</button>
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