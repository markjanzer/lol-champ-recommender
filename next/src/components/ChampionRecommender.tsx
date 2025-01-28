'use client';
import { useEffect, useState } from "react";
import { recommendChampions } from "@/lib/champions/recommendationEngine";
import { ChampionDataMap, ChampionPerformance, Champion } from "@/lib/types/champions"
import ChampionCombobox from "./ChampionCombobox";
import ChampionRecommendation from "./ChampionRecommendation";
import * as Tooltip from "@radix-ui/react-tooltip";
import QuestionMarkIcon from "./QuestionMarkIcon";

interface Props {
  championStats: ChampionDataMap;
  champions: Champion[];
}

function renderTooltip() {
  return (
    <Tooltip.Provider delayDuration={100}>
      <Tooltip.Root>
        <Tooltip.Trigger>
          <QuestionMarkIcon />
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

  const notEmpty = allies.some(ally => ally !== null) || enemies.some(enemy => enemy !== null) || bans.some(ban => ban !== null);

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
        {
          notEmpty && (
            <div className="mt-8 flex justify-center">
              <button className="bg-gray-200 dark:bg-gray-800 px-4 py-2 rounded-md" onClick={clearAll}>Clear</button>
            </div>
          )
        }
        <div className="my-8 flex flex-col items-center">
          <h2 className="text-xl font-bold text-center mb-2">About Champ Recs</h2>
          <p className="text-md max-w-prose">
            This tool helps you make smarter champion selections in League of Legends by analyzing how well each potential pick performs with allies and against enemies. Drawing from over 100,000 ranked games, it calculates champion synergies and counters irrespective of role.
          </p>
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