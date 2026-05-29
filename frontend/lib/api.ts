export type Account = {
  id: string;
  provider: string;
  name: string;
  type: string;
  currency: string;
  balancePence: number;
};

export type Transaction = {
  id: string;
  accountId: string;
  amountPence: number;
  currency: string;
  description: string;
  merchantName?: string;
  occurredAt: string;
};

const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080';

export async function fetchAccounts(): Promise<Account[]> {
  return fetchJson<Account[]>('/api/accounts');
}

export async function fetchTransactions(): Promise<Transaction[]> {
  return fetchJson<Transaction[]>('/api/transactions');
}

async function fetchJson<T>(path: string): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, { cache: 'no-store' });

  if (!response.ok) {
    throw new Error(`PennyPilot API request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}
