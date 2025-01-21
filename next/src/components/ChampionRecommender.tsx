'use client';
import { useEffect, useState } from "react";
import { recommendChampions } from "@/lib/champions/recommendationEngine";
import { ChampionDataMap, ChampionPerformance, Champion } from "@/lib/types/champions"
import ChampionCombobox from "./ChampionCombobox";
import ChampionRecommendation from "./ChampionRecommendation";
import * as Tooltip from "@radix-ui/react-tooltip";

interface Props {
  championStats: ChampionDataMap;
  champions: Champion[];
}

function tooltipIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 15 15" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M0.877075 7.49972C0.877075 3.84204 3.84222 0.876892 7.49991 0.876892C11.1576 0.876892 14.1227 3.84204 14.1227 7.49972C14.1227 11.1574 11.1576 14.1226 7.49991 14.1226C3.84222 14.1226 0.877075 11.1574 0.877075 7.49972ZM7.49991 1.82689C4.36689 1.82689 1.82708 4.36671 1.82708 7.49972C1.82708 10.6327 4.36689 13.1726 7.49991 13.1726C10.6329 13.1726 13.1727 10.6327 13.1727 7.49972C13.1727 4.36671 10.6329 1.82689 7.49991 1.82689ZM8.24993 10.5C8.24993 10.9142 7.91414 11.25 7.49993 11.25C7.08571 11.25 6.74993 10.9142 6.74993 10.5C6.74993 10.0858 7.08571 9.75 7.49993 9.75C7.91414 9.75 8.24993 10.0858 8.24993 10.5ZM6.05003 6.25C6.05003 5.57211 6.63511 4.925 7.50003 4.925C8.36496 4.925 8.95003 5.57211 8.95003 6.25C8.95003 6.74118 8.68002 6.99212 8.21447 7.27494C8.16251 7.30651 8.10258 7.34131 8.03847 7.37854L8.03841 7.37858C7.85521 7.48497 7.63788 7.61119 7.47449 7.73849C7.23214 7.92732 6.95003 8.23198 6.95003 8.7C6.95004 9.00376 7.19628 9.25 7.50004 9.25C7.8024 9.25 8.04778 9.00601 8.05002 8.70417L8.05056 8.7033C8.05924 8.6896 8.08493 8.65735 8.15058 8.6062C8.25207 8.52712 8.36508 8.46163 8.51567 8.37436L8.51571 8.37433C8.59422 8.32883 8.68296 8.27741 8.78559 8.21506C9.32004 7.89038 10.05 7.35382 10.05 6.25C10.05 4.92789 8.93511 3.825 7.50003 3.825C6.06496 3.825 4.95003 4.92789 4.95003 6.25C4.95003 6.55376 5.19628 6.8 5.50003 6.8C5.80379 6.8 6.05003 6.55376 6.05003 6.25Z" fill="currentColor" fill-rule="evenodd" clip-rule="evenodd"></path></svg>
  )
}

function renderTooltip() {
  return (
    <Tooltip.Provider delayDuration={100}>
      <Tooltip.Root>
        <Tooltip.Trigger>
          {tooltipIcon()}
        </Tooltip.Trigger>
        <Tooltip.Portal>
          <Tooltip.Content side="right" sideOffset={8} className="max-w-[250px]">
            <div className="bg-gray-200 rounded-md p-2 text-sm">
              Average of the champion&apos;s synergy and matchup winrates.
            </div>
          </Tooltip.Content>
        </Tooltip.Portal>
      </Tooltip.Root>
    </Tooltip.Provider>
  )
}

export default function ChampionRecommender({championStats, champions}: Props) {
  const TEAM_SIZE = 5;
  const BANS_SIZE = 10;
  
  const [allies, setAllies] = useState<(Champion | null)[]>(Array(TEAM_SIZE).fill(null));
  const [enemies, setEnemies] = useState<(Champion | null)[]>(Array(TEAM_SIZE).fill(null));
  const [bans, setBans] = useState<(Champion | null)[]>(Array(BANS_SIZE).fill(null));
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

  const clearAll = () => {
    setAllies(Array(TEAM_SIZE).fill(null));
    setEnemies(Array(TEAM_SIZE).fill(null));
    setBans(Array(BANS_SIZE).fill(null));
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

  return (
    <div className="grid md:grid-cols-5">
      <div className="md:col-span-3">
        <h1 className="text-2xl font-bold text-center">Selected Champions</h1>
        <div className="mt-4 flex flex-row justify-around">
          <div className="flex flex-col">
            <h2 className="text-lg font-bold text-center">Allies</h2>
            {allies.map((ally, index) => (
              <ChampionCombobox 
                key={index} 
                champions={champions} 
                onChange={(champion) => handleAllyChange(index, champion)} 
                value={ally} 
              />
            ))}
          </div>
          <div className="flex flex-col">
            <h2 className="text-lg font-bold text-center">Enemies</h2>
            {enemies.map((enemy, index) => (
              <ChampionCombobox 
                key={index} 
                champions={champions} 
                onChange={(champion) => handleEnemyChange(index, champion)} 
                value={enemy} 
              />
            ))}
          </div>
        </div>
        <div className="mt-8">
          <h2 className="text-lg font-bold text-center">Bans</h2>
          <div className="flex flex-row justify-around">
            <div className="flex flex-col">
              {bans.slice(0, 5).map((bannedChampion, index) => (
                <ChampionCombobox 
                  key={index} 
                  champions={champions} 
                  onChange={(champion) => handleBanChange(index, champion)} 
                  value={bannedChampion} 
                />
              ))}
            </div>
            <div className="flex flex-col">
              {bans.slice(5, 10).map((bannedChampion, index) => (
                <ChampionCombobox 
                  key={index + 5} 
                  champions={champions} 
                  onChange={(champion) => handleBanChange(index + 5, champion)} 
                  value={bannedChampion} 
                />
              ))}
            </div>
          </div>
        </div>
        <div className="mt-8 flex justify-center">
          <button className="bg-gray-200 px-4 py-2 rounded-md" onClick={clearAll}>Clear</button>
        </div>
      </div>
      <div className="md:col-span-2 md:border-l mt-8 md:mt-0 px-8">
        <h2 className="text-2xl font-bold">Recommended Champions {renderTooltip()}</h2>
        <div className="mt-4">
          {recommendations.map((recommendation) => (
            <ChampionRecommendation 
              key={recommendation.championId} 
              championPerformance={recommendation} 
              champions={champions}
            />
          ))}
        </div>
      </div>
    </div>
  )
}