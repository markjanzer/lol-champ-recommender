'use client';
import { useEffect, useState, Fragment } from "react";

interface Props {
  champions: {
    name: string;
    api_id: number;
  }[];
}

export default function ChampionCombobox({champions}: Props) {
  return (
    <select className="border border-gray-300 rounded-md p-2 text-black">
      {champions.map(champion => (
        <option key={champion.api_id} value={champion.api_id}>{champion.name}</option>
      ))}
    </select>
  );
}