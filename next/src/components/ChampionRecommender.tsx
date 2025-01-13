'use client';
import { useState } from "react";

export default function ChampionRecommender() {
  const [allies, setAllies] = useState<number[]>([0,0,0,0]);
  const [enemies, setEnemies] = useState<number[]>([0,0,0,0,0]);

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
          <p>Allies: {allies.join(', ')}</p>
          <p>Enemies: {enemies.join(', ')}</p>
        </div>
      </div>
    </>
  )
}