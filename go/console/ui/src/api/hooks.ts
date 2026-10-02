import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { api } from './client';
import type {
  Catalog,
  CreateShopRequest,
  InboundFilter,
  InboundLog,
  InboundLogSummary,
  LoginRequest,
  LoginResponse,
  MeResponse,
  OutboundFilter,
  OutboundLog,
  OutboundLogSummary,
  Paginated,
  SaveGoldenRequest,
  SaveGoldenResponse,
  SendRequestBody,
  Shop,
  Stats,
  UpdateShopRequest,
  VerifyShopResponse,
} from './types';

/** Query keys shared by hooks and invalidations. */
export const queryKeys = {
  me: ['me'] as const,
  shops: ['shops'] as const,
  shop: (id: number) => ['shops', id] as const,
  catalog: ['catalog'] as const,
  outbound: (filter: OutboundFilter) => ['outbound', filter] as const,
  outboundDetail: (id: number) => ['outbound', 'detail', id] as const,
  inbound: (filter: InboundFilter) => ['inbound', filter] as const,
  inboundDetail: (id: number) => ['inbound', 'detail', id] as const,
  stats: ['stats'] as const,
};

/** Current user and setup flag; 40100 does not redirect (the guard decides). */
export function useMe(): UseQueryResult<MeResponse> {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: () => api.get<MeResponse>('/api/auth/me', undefined, true),
    retry: false,
    staleTime: 60_000,
  });
}

/** Logs in and refreshes the `me` query. */
export function useLogin(): UseMutationResult<LoginResponse, Error, LoginRequest> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: LoginRequest) => api.post<LoginResponse>('/api/auth/login', body),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: queryKeys.me });
    },
  });
}

/** Logs out and clears every cached query. */
export function useLogout(): UseMutationResult<Record<string, never>, Error, void> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<Record<string, never>>('/api/auth/logout'),
    onSuccess: () => {
      qc.clear();
    },
  });
}

/** All Shops. */
export function useShops(): UseQueryResult<Shop[]> {
  return useQuery({ queryKey: queryKeys.shops, queryFn: () => api.get<Shop[]>('/api/shops') });
}

/** Invalidates the queries affected by a Shop change. */
function useInvalidateShops(): () => Promise<void> {
  const qc = useQueryClient();
  return async () => {
    await Promise.all([
      qc.invalidateQueries({ queryKey: queryKeys.shops }),
      qc.invalidateQueries({ queryKey: queryKeys.me }),
      qc.invalidateQueries({ queryKey: queryKeys.stats }),
    ]);
  };
}

/** Creates a Shop (the backend validates it against CYBERBIZ). */
export function useCreateShop(): UseMutationResult<Shop, Error, CreateShopRequest> {
  const invalidate = useInvalidateShops();
  return useMutation({
    mutationFn: (body: CreateShopRequest) => api.post<Shop>('/api/shops', body),
    onSuccess: invalidate,
  });
}

/** Variables for {@link useUpdateShop}. */
export interface UpdateShopVariables {
  id: number;
  body: UpdateShopRequest;
}

/** Updates a Shop; blank secret/token keep the stored ones. */
export function useUpdateShop(): UseMutationResult<Shop, Error, UpdateShopVariables> {
  const invalidate = useInvalidateShops();
  return useMutation({
    mutationFn: ({ id, body }: UpdateShopVariables) => api.put<Shop>(`/api/shops/${id}`, body),
    onSuccess: invalidate,
  });
}

/** Deletes a Shop; its logs are kept. */
export function useDeleteShop(): UseMutationResult<Record<string, never>, Error, number> {
  const invalidate = useInvalidateShops();
  return useMutation({
    mutationFn: (id: number) => api.delete<Record<string, never>>(`/api/shops/${id}`),
    onSuccess: invalidate,
  });
}

/** Calls `GET /shop` through the SDK for a Shop and returns the reply. */
export function useVerifyShop(): UseMutationResult<VerifyShopResponse, Error, number> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.post<VerifyShopResponse>(`/api/shops/${id}/verify`),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ['outbound'] });
    },
  });
}

/** The OpenAPI catalog used by the API tester. Rarely changes: cached for the session. */
export function useCatalog(): UseQueryResult<Catalog> {
  return useQuery({
    queryKey: queryKeys.catalog,
    queryFn: () => api.get<Catalog>('/api/catalog'),
    staleTime: Infinity,
  });
}

/** Variables for {@link useSendRequest}. */
export interface SendRequestVariables {
  shopId: number;
  body: SendRequestBody;
}

/** Executes an Outbound through a Shop's SDK client and returns the recorded log. */
export function useSendRequest(): UseMutationResult<OutboundLog, Error, SendRequestVariables> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ shopId, body }: SendRequestVariables) =>
      api.post<OutboundLog>(`/api/shops/${shopId}/requests`, body),
    onSuccess: async () => {
      await Promise.all([
        qc.invalidateQueries({ queryKey: ['outbound'] }),
        qc.invalidateQueries({ queryKey: queryKeys.stats }),
      ]);
    },
  });
}

/** Paginated Outbound list. */
export function useOutbound(filter: OutboundFilter): UseQueryResult<Paginated<OutboundLogSummary>> {
  return useQuery({
    queryKey: queryKeys.outbound(filter),
    queryFn: () => api.get<Paginated<OutboundLogSummary>>('/api/outbound', { ...filter }),
    placeholderData: (prev) => prev,
  });
}

/** One Outbound with bodies; disabled while `id` is null. */
export function useOutboundDetail(id: number | null): UseQueryResult<OutboundLog> {
  return useQuery({
    queryKey: queryKeys.outboundDetail(id ?? 0),
    queryFn: () => api.get<OutboundLog>(`/api/outbound/${id ?? 0}`),
    enabled: id !== null,
  });
}

/** Paginated Inbound list. */
export function useInbound(filter: InboundFilter): UseQueryResult<Paginated<InboundLogSummary>> {
  return useQuery({
    queryKey: queryKeys.inbound(filter),
    queryFn: () => api.get<Paginated<InboundLogSummary>>('/api/inbound', { ...filter }),
    placeholderData: (prev) => prev,
  });
}

/** One Inbound with headers and body; disabled while `id` is null. */
export function useInboundDetail(id: number | null): UseQueryResult<InboundLog> {
  return useQuery({
    queryKey: queryKeys.inboundDetail(id ?? 0),
    queryFn: () => api.get<InboundLog>(`/api/inbound/${id ?? 0}`),
    enabled: id !== null,
  });
}

/** Which log kind a Golden File is written from. */
export type GoldenKind = 'outbound' | 'inbound';

/** Variables for {@link useSaveGolden}. */
export interface SaveGoldenVariables {
  id: number;
  body: SaveGoldenRequest;
}

/** Writes a Golden File from an Outbound or Inbound. */
export function useSaveGolden(
  kind: GoldenKind,
): UseMutationResult<SaveGoldenResponse, Error, SaveGoldenVariables> {
  return useMutation({
    mutationFn: ({ id, body }: SaveGoldenVariables) =>
      api.post<SaveGoldenResponse>(`/api/${kind}/${id}/golden`, body),
  });
}

/** Dashboard counters, refreshed every 30 seconds. */
export function useStats(): UseQueryResult<Stats> {
  return useQuery({
    queryKey: queryKeys.stats,
    queryFn: () => api.get<Stats>('/api/stats'),
    refetchInterval: 30_000,
  });
}
