import { useUser } from "@/app/components/provider/UserProvider";
import { addHfToken as apiAdd, deleteHfToken as apiDelete } from "@/app/components/handlers/tokens";
import toast from "react-hot-toast";
import { User } from "@/app/components/handlers/auth";

export const useHfTokens = () => {
  const { user, setUser } = useUser();

  const addToken = async (name: string, token: string) => {
    if (!name.trim() || !token.trim()) return [];

    const updatedTokenNames = await apiAdd(name.trim(), token.trim());

    setUser((prev: User | null) =>
      prev
        ? {
            ...prev,
            hf_token_names: updatedTokenNames,
          }
        : null,
    );

    return updatedTokenNames;
  };

  const deleteToken = async (name: string) => {
    if (!name.trim()) return [];

    const updatedTokenNames = await apiDelete(name.trim());

    setUser((prev: User | null) =>
      prev
        ? {
            ...prev,
            hf_token_names: updatedTokenNames,
          }
        : null,
    );

    return updatedTokenNames;
  };

  const listHfTokens = () => user?.hf_token_names ?? [];

  return {
    user,
    addToken,
    deleteToken,
    listHfTokens,
  };
};
