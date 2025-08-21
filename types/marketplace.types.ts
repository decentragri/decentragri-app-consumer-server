export enum ListingStatus {
  UNSET = 'UNSET',
  CREATED = 'CREATED',
  COMPLETED = 'COMPLETED',
  CANCELLED = 'CANCELLED',
  ACTIVE = 'ACTIVE',
  EXPIRED = 'EXPIRED'
}

export interface CurrencyValuePerToken {
  name: string;
  symbol: string;
  decimals: number;
  value: string;
  displayValue: string;
}

export interface DirectListing {
  id: string;
  assetContractAddress: string;
  tokenId: string;
  seller?: string;
  pricePerToken: string;
  currencyContractAddress: string;
  quantity: string;
  isReservedListing: boolean;
  currencyValuePerToken: CurrencyValuePerToken;
  startTimeInSeconds: number;
  endTimeInSeconds: number;
  status: ListingStatus;
}

export interface Coordinates {
  lat: number;
  long: number;
}

export interface FarmPlotAttributes {
  id: string;
  price: string;
  farmName: string;
  description: string;
  cropType: string;
  owner: string;
  image: string;
  location: string;
  coordinates: Coordinates;
  createdAt: string;
}

export interface FarmPlotMetadata {
  name: string;
  description?: string;
  image?: string;
  external_url?: string;
  background_color?: string;
  properties?: Record<string, any>;
  attributes?: FarmPlotAttributes[];
}

export interface FarmPlotDirectListingWithImageByte extends DirectListing {
  asset: FarmPlotMetadata;
  imageBytes: number[] | null; // Array of uint8 values
}

// This is the response type for GetAllValidFarmPlotListings
export type GetAllValidFarmPlotListingsResponse = FarmPlotDirectListingWithImageByte[];

// Example usage:
/*
async function getAllValidFarmPlotListings(): Promise<GetAllValidFarmPlotListingsResponse> {
  const response = await fetch('/api/marketplace/valid-farmplots', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch farm plot listings');
  }

  return response.json();
}
*/
