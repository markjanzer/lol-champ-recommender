'use client';
import { useEffect, useState, Fragment } from "react";
import { recommendChampions } from "@/lib/champions/recommendationEngine";
import { ChampionDataMap, ChampionPerformance, Champion } from "@/lib/types/champions"
import ChampionCombobox from "./ChampionCombobox";

interface Props {
  championStats: ChampionDataMap;
  champions: Champion[];
}

export default function ChampionRecommender({championStats, champions}: Props) {
  const [allies, setAllies] = useState<(Champion | null)[]>([null, null, null, null, null]);
  const [enemies, setEnemies] = useState<(Champion | null)[]>([null, null, null, null, null]);
  const [bans, setBans] = useState<(Champion | null)[]>([null, null, null, null, null, null, null, null, null, null]);
  const [recommendations, setRecommendations] = useState<ChampionPerformance[]>([]);

  const handleAllyChange = (index: number, value: Champion) => {
    const newAllies = [...allies];
    newAllies[index] = value;
    setAllies(newAllies);
  };

  const handleEnemyChange = (index: number, value: Champion) => {
    const newEnemies = [...enemies];
    newEnemies[index] = value;
    setEnemies(newEnemies);
  };

  const handleBanChange = (index: number, value: Champion) => {
    const newBans = [...bans];
    newBans[index] = value;
    setBans(newBans);
  };

  useEffect(() => {
    const validAllies = allies.filter((ally): ally is Champion => ally !== null);
    const validEnemies = enemies.filter((enemy): enemy is Champion => enemy !== null);
    const validBans = bans.filter((ban): ban is Champion => ban !== null);

    if (validAllies.length === 0 && validEnemies.length === 0) {
      setRecommendations([]);
      return;
    }
    
    const recommendations = recommendChampions(championStats, { 
      allies: validAllies.map(ally => ally.api_id), 
      enemies: validEnemies.map(enemy => enemy.api_id), 
      bans: validBans.map(ban => ban.api_id)
    });
    setRecommendations(recommendations);
  }, [championStats, champions, allies, enemies, bans]);


  const getChampionName = (apiId: number) => {
    return champions.find(champion => champion.api_id === apiId)?.name;
  };

  const formatWinrate = (winrate: number) => {
    return `${(winrate * 100).toFixed(2)}%`;
  };

  return (
    <>
      <div>
        <h1 className="text-2xl font-bold">Pick Champions</h1>
        <div className="mt-4 grid grid-cols-3">
          <div className="col-span-1">
            <h2 className="text-lg font-bold">Allies</h2>
            {allies.map((ally, index) => (
              <ChampionCombobox 
                key={index} 
                champions={champions} 
                onChange={(champion) => handleAllyChange(index, champion)} 
                value={ally} 
              />
            ))}
          </div>
          <div className="col-span-1">
            <h2 className="text-lg font-bold">Enemies</h2>
            {enemies.map((enemy, index) => (
              <ChampionCombobox 
                key={index} 
                champions={champions} 
                onChange={(champion) => handleEnemyChange(index, champion)} 
                value={enemy} 
              />
            ))}
          </div>
          <div className="col-span-1">
            <h2 className="text-lg font-bold">Bans</h2>
            {bans.map((bannedChampion, index) => (
              <ChampionCombobox 
                key={index} 
                champions={champions} 
                onChange={(champion) => handleBanChange(index, champion)} 
                value={bannedChampion} 
              />
            ))}
          </div>
        </div>
      </div>
      <div className="mt-4">
        <h2 className="text-2xl font-bold">Recommended Champions</h2>
        <div className="mt-4">
          {recommendations.map((recommendation) => (
            <div key={recommendation.championId}>
              <p className="text-lg font-bold">{getChampionName(recommendation.championId)}: {formatWinrate(recommendation.winProbability)}</p>
              <div>
                <p>Synergies: [{recommendation.synergies.map(synergy => `${getChampionName(synergy.championId)} - ${formatWinrate(synergy.winProbability)}`).join(', ')}]</p>
                <p>Matchups: [{recommendation.matchups.map(matchup => `${getChampionName(matchup.championId)} - ${formatWinrate(matchup.winProbability)}`).join(', ')}]</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  )
}