import React, { useState, useEffect } from "react";
import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from "axios";
import user from "./user.json";
const url = "https://jsonplaceholder.typicode.com";

type USER = typeof user;

// const options: AxiosRequestConfig = {
const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};

// const getUsersInfo = async (): Promise<USERS> => {
const getUsersInfo = async (): Promise<USER[]> => {
  // const getUsersInfo = async (): USERS => {
  const res: AxiosResponse<USER[]> = await axios(options);
  const { data } = res;

  // const { data } = await axios.get<USER[]>(`${url}/users`);

  return data;
};

const AxiosResReqType: React.FC = () => {
  const [users, setUsers] = useState<USER[]>([]);
  const [status, setStatus] = useState<number | null>(null);

  useEffect(() => {
    axios.get<USER[]>(`${url}/users`).then((res) => {
      const { data, status } = res;
      setUsers(data);
      setStatus(status);
    });

    axios(options)
      .then((res: AxiosResponse<USER[]>) => {
        const { data, status } = res;
        setUsers(data);
        setStatus(status);
      })
      .catch((e: AxiosError<{ error: string }>) => {
        // エラー処理
        console.log(e.message);
      });

    axios(options).then(
      ({ data, status }: { data: USER[]; status: number }) => {
        setUsers(data);
        setStatus(status);
      }
    );

    type resType = { data: USER[]; status: number };
    axios(options).then(({ data, status }: resType) => {
      setUsers(data);
      setStatus(status);
    });
  }, []);
  return (
    <div>
      <h1>Axiosのリクエストとレスポンスに型をつける</h1>
      <h2>ステータス:{status}</h2>

      <ul>
        {users.map(({ id, name }) => {
          return (
            <li key={id}>
              {id} : {name}
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default AxiosResReqType;
