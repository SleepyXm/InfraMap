import { useState, useEffect } from "react";
import * as d3 from "d3";
import * as topojson from "topojson-client";

export function useWorldMap(dims: { w: number; h: number }) {
  const [worldData, setWorldData] = useState<any>(null);

  useEffect(() => {
    fetch("https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json")
      .then(r => r.json())
      .then(setWorldData);
  }, []);

  if (!worldData || dims.w === 0) return null;

  const proj = d3.geoNaturalEarth1()
    .scale(dims.w / 6.3)
    .translate([dims.w / 2, dims.h / 2.1]);

  const pathGen = d3.geoPath().projection(proj);
  const land = topojson.feature(worldData, worldData.objects.land);
  const borders = topojson.mesh(worldData, worldData.objects.countries, (a: any, b: any) => a !== b);
  const projectLatLng = (lat: number, lng: number): [number, number] | null => proj([lng, lat]) ?? null;

  return { pathGen, land, borders, projectLatLng };
}