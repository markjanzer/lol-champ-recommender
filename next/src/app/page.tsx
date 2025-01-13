import db from "@/lib/db";

export default async function Home() {
  const champions = await db.query('SELECT name, api_id FROM champions ORDER BY name ASC');

  return (
    <div className="grid grid-cols-4 mx-auto max-w-6xl mt-10">
      <div className="col-span-1">
        <h2 className="text-2xl font-bold">Champions</h2>
        <table className="table-fixed">
          <tbody>
            {champions.rows.map((champion) => (
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
        <div>
          <h1 className="text-2xl font-bold">Pick Champions</h1>
          <div className="mt-4">
            <h2 className="text-lg font-bold">Allies</h2>
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
          </div>
          <div className="mt-4">
            <h2 className="text-lg font-bold">Enemies</h2>
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
            <input className="border border-gray-300 rounded-md p-2" type="number" min={1} max={10000} />
          </div>
          <div className="mt-4">
            <button className="bg-blue-500 text-white rounded-md p-2">Recommend</button>
          </div>
        </div>
        <div className="mt-4">
          <h2 className="text-2xl font-bold">Recommended Champions</h2>
        </div>
      </div>
    </div>
  );
}
